package upgrader

import (
	"fmt"
	"log"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/workers"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CCAPI
type CCAPI interface {
	GetServiceInstances([]string) ([]ccapi.ServiceInstance, error)
	GetServicePlans(string) ([]ccapi.Plan, error)
	UpgradeServiceInstance(string, string) error
}

func Upgrade(api CCAPI, brokerName string, batchSize int) error {
	plans, err := api.GetServicePlans(brokerName)
	if err != nil {
		return err
	}

	var planGUIDS []string
	planVersions := make(map[string]string)

	for _, plan := range plans {
		planGUIDS = append(planGUIDS, plan.GUID)
		planVersions[plan.GUID] = plan.MaintenanceInfoVersion
	}

	fmt.Printf("Discovering service instances for broker: %s\n", brokerName)

	serviceInstances, err := api.GetServiceInstances(planGUIDS)
	if err != nil {
		return err
	}

	type upgradeTask struct {
		Guid      string
		MIVersion string
	}

	logUpgradableInstances(serviceInstances)

	upgradeQueue := make(chan upgradeTask)

	fmt.Printf("Starting upgrade\n")

	for i := range planVersions {
		fmt.Printf("Plan version: %s\n", i)
	}

	for _, ins := range serviceInstances {
		fmt.Printf("service version: %s\n", ins.PlanGUID)
	}

	go func() {
		for _, instance := range serviceInstances {
			if instance.UpgradeAvailable {
				fmt.Printf("Instance PlanGUID: %s\n", instance.PlanGUID)
				upgradeQueue <- upgradeTask{
					Guid:      instance.GUID,
					MIVersion: planVersions[instance.PlanGUID],
				}
			}
		}
		close(upgradeQueue)
	}()

	workers.Run(batchSize, func() {
		for instance := range upgradeQueue {
			err := api.UpgradeServiceInstance(instance.Guid, instance.MIVersion)
			if err != nil {
				fmt.Printf("error upgrading service instance: %s\n", err)
			}
		}
	})

	return nil
}

func logUpgradableInstances(si []ccapi.ServiceInstance) {
	upgradableInstances := 0
	for _, i := range si {
		if i.UpgradeAvailable {
			upgradableInstances++
		}
	}

	log.Printf("---\n"+
		"Total instances: %d\n"+
		"Upgradable instances: %d\n"+
		"---\n",
		len(si),
		upgradableInstances)
}
