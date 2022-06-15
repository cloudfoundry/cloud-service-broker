package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/cli/plugin"
)

func main() {
	plugin.Start(new(UpgradePlugin))
}

func runUpgrade(apiToken, apiURL, brokerName string) {
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

	type upgradeTask struct {
		serviceInstanceGUID string
		newVersion          string
	}

	upgradeQueue := make(chan upgradeTask)

	go func() {
		for _, instance := range instances {
			if instance.UpgradeAvailable {
				newVersion := planVersions[instance.Relationships.ServicePlan.Data.GUID]
				upgradeQueue <- upgradeTask{
					serviceInstanceGUID: instance.GUID,
					newVersion:          newVersion,
				}
			}
		}
		close(upgradeQueue)
	}()

	log.Printf("---\n"+
		"Total instances: %d\n"+
		"Upgradable instances: %d\n"+
		"---\n",
		len(instances),
		len(upgradeQueue))

	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)

	log.Printf("Starting Upgrade\n")

	var failedInstances []string

	for i := 0; i < workers; i++ {
		go func() {
			for task := range upgradeQueue {
				err := upgrade(r, task.serviceInstanceGUID, task.newVersion)
				if err != nil {
					failedInstances = append(failedInstances, task.serviceInstanceGUID)
				}
			}
			wg.Done()
		}()
	}

	totalInstances := len(instances)

	for range upgradeQueue {
		log.Printf("Upgraded %d/%d\n", totalInstances-len(upgradeQueue), totalInstances)
	}

	wg.Wait()

	log.Printf("---\n"+
		"Finished upgrade:\n"+
		"Total instances upgraded: %d\n"+
		"Failed to upgrade instances:\n %s\n"+
		"---\n",
		totalInstances-len(failedInstances),
		strings.Join(failedInstances, "\n"))

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
		Type  string `json:"type"`
		State string `json:"state"`
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

		if receiver.LastOperation.Type == "update" && receiver.LastOperation.State == "failed" {
			return fmt.Errorf("failed to update ")
		}

		if receiver.LastOperation.Type != "update" || receiver.LastOperation.State != "in progress" {
			return nil
		}
		time.Sleep(time.Second)
	}
}
