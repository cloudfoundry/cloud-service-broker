package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/cli/plugin"
)

func main() {
	plugin.Start(new(UpgradePlugin))
}

func runUpgrade(apiToken, apiURL, brokerName string) {

	r := NewRequester(apiURL, apiToken)

	plans := plansForBroker(r, brokerName)

	var planGUIDs []string
	planVersions := make(map[string]string)

	for _, p := range plans {
		planGUIDs = append(planGUIDs, p.GUID)
		planVersions[p.GUID] = p.MaintenanceInfo.Version
	}

	instances := serviceInstances(r, planGUIDs)

	type upgradeTask struct {
		serviceInstanceGUID string
		newVersion          string
	}

	work := make(chan upgradeTask)
	go func() {
		for _, instance := range instances {
			if instance.UpgradeAvailable {
				newVersion := planVersions[instance.Relationships.ServicePlan.Data.GUID]
				fmt.Printf("Will upgrade %s to %s\n", instance.GUID, newVersion)
				work <- upgradeTask{
					serviceInstanceGUID: instance.GUID,
					newVersion:          newVersion,
				}
			}
		}
		close(work)
	}()

	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			for task := range work {
				upgrade(r, task.serviceInstanceGUID, task.newVersion)
			}
			wg.Done()
		}()
	}

	wg.Wait()
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

func upgrade(r Requester, serviceInstanceGUID, newVersion string) {
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

		if receiver.LastOperation.Type != "update" || receiver.LastOperation.State != "in progress" {
			return
		}
		time.Sleep(time.Second)
	}
}
