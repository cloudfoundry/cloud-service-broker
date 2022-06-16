package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/cli/plugin"
)

func main() {
	plugin.Start(new(UpgradePlugin))
}

func runUpgrade(apiToken, apiURL, brokerName string, dryRun bool) {
	log.SetOutput(os.Stdout)

	r := NewRequester(apiURL, apiToken)

	plans := plansForBroker(r, brokerName)

	var planGUIDs []string
	planVersions := make(map[string]string)

	for _, p := range plans {
		planGUIDs = append(planGUIDs, p.GUID)
		planVersions[p.GUID] = p.MaintenanceInfo.Version
	}

	log.Printf("Discovering service instances for broker: %s", brokerName)

	instances := serviceInstances(r, planGUIDs)

	var upgradableInstances int
	for _, instance := range instances {
		if instance.UpgradeAvailable {
			upgradableInstances++
		}
	}

	log.Printf("---\n"+
		"Total instances: %d\n"+
		"Upgradable instances: %d\n"+
		"---\n",
		len(instances),
		upgradableInstances)

	if dryRun {
		return
	}

	type upgradeTask struct {
		serviceInstanceGUID string
		newVersion          string
	}

	upgradeQueue := make(chan upgradeTask)

	go func() {
		for _, instance := range instances {
			if instance.UpgradeAvailable {
				log.Printf("Instance upgradable: %v", instance.GUID)
				newVersion := planVersions[instance.Relationships.ServicePlan.Data.GUID]
				upgradeQueue <- upgradeTask{
					serviceInstanceGUID: instance.GUID,
					newVersion:          newVersion,
				}
			}
		}
		log.Printf("closing upgradeQueue")
		close(upgradeQueue)
	}()

	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)

	log.Printf("Starting Upgrade\n")

	var failedInstances map[string]string
	addFailedInstance := func(instance, description string) {
		var lock sync.Mutex

		lock.Lock()
		defer lock.Unlock()

		failedInstances[instance] = description
	}

	var upgraded int32
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for task := range upgradeQueue {
				err := upgrade(r, task.serviceInstanceGUID, task.newVersion)
				if err != nil {
					addFailedInstance(task.serviceInstanceGUID, err.Error())
				}
				atomic.AddInt32(&upgraded, 1)
			}
			log.Printf("worker done")
		}()
	}

	for range upgradeQueue {
		log.Printf("Upgraded %d/%d\n", upgraded, upgradableInstances)
		time.Sleep(30 * time.Second)
	}

	wg.Wait()

	log.Printf("---\n"+
		"Finished upgrade:\n"+
		"Total instances upgraded: %d\n",
		upgraded)

	if len(failedInstances) > 0 {
		fmt.Printf(
			"Failed to upgrade instances:\n" +
				"GUID\tError\n")

		for k, v := range failedInstances {
			fmt.Printf("%s\t%s\n", k, v)
		}
	}

}

type plan struct {
	GUID            string `json:"guid"`
	MaintenanceInfo struct {
		Version string `json:"version"`
	} `json:"maintenance_info"`
}

func plansForBroker(r Requester, brokerName string) []plan {
	var receiver struct {
		Resources []plan `json:"resources"`
	}
	r.Get(fmt.Sprintf("v3/service_plans?per_page=5000&service_broker_names=%s", brokerName), &receiver)

	return receiver.Resources
}

type serviceInstance struct {
	GUID             string `json:"guid"`
	UpgradeAvailable bool   `json:"upgrade_available"`
	Relationships    struct {
		ServicePlan struct {
			Data struct {
				GUID string `json:"guid"`
			} `json:"data"`
		} `json:"service_plan"`
	} `json:"relationships"`

	LastOperation struct {
		Type        string `json:"type"`
		State       string `json:"state"`
		Description string `json:"description"`
	} `json:"last_operation"`
}

func serviceInstances(r Requester, planGUIDs []string) []serviceInstance {
	var receiver struct {
		Resources []serviceInstance `json:"resources"`
	}
	r.Get(fmt.Sprintf("v3/service_instances?per_page=5000&service_plan_guids=%s", strings.Join(planGUIDs, ",")), &receiver)

	return receiver.Resources
}

func upgrade(r Requester, serviceInstanceGUID, newVersion string) error {
	var body struct {
		MaintenanceInfo struct {
			Version string `json:"version"`
		} `json:"maintenance_info"`
	}
	body.MaintenanceInfo.Version = newVersion
	r.Patch(fmt.Sprintf("v3/service_instances/%s", serviceInstanceGUID), body)

	for {
		var receiver serviceInstance
		r.Get(fmt.Sprintf("v3/service_instances/%s", serviceInstanceGUID), &receiver)

		if receiver.LastOperation.State == "failed" {
			return fmt.Errorf(receiver.LastOperation.Description)
		}

		if receiver.LastOperation.Type != "update" || receiver.LastOperation.State != "in progress" {
			return nil
		}
		time.Sleep(time.Second)
	}
}
