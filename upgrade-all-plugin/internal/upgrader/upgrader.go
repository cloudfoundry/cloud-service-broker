package upgrader

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

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

	upgradableInstances := 0
	for _, i := range serviceInstances {
		if i.UpgradeAvailable {
			upgradableInstances++
		}
	}

	fmt.Printf("---\n"+
		"Total instances: %d\n"+
		"Upgradable instances: %d\n"+
		"---\n",
		len(serviceInstances),
		upgradableInstances)

	upgradeQueue := make(chan upgradeTask)

	fmt.Printf("Starting upgrade...\n")
	go func() {
		for _, instance := range serviceInstances {
			if instance.UpgradeAvailable {
				upgradeQueue <- upgradeTask{
					Guid:      instance.GUID,
					MIVersion: planVersions[instance.PlanGUID],
				}
			}
		}
		close(upgradeQueue)
	}()

	var failedInstances map[string]string
	addFailedInstance := func(instance, description string) {
		var lock sync.Mutex

		lock.Lock()
		defer lock.Unlock()

		failedInstances[instance] = description
	}

	var upgraded int32

	upgradeComplete := make(chan bool)

	go func() {
		for {
			select {
			case <-upgradeComplete:
				return
			case <-time.After(time.Minute):
				fmt.Printf("Upgraded %d/%d\n", upgraded, upgradableInstances)
			}
		}
	}()

	workers.Run(batchSize, func() {
		for instance := range upgradeQueue {
			err := api.UpgradeServiceInstance(instance.Guid, instance.MIVersion)
			if err != nil {
				fmt.Printf("error upgrading service instance: %s\n", err)
				addFailedInstance(instance.Guid, err.Error())
			}
			atomic.AddInt32(&upgraded, 1)
		}
	})

	upgradeComplete <- true

	fmt.Printf("---\n"+
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

	return nil
}

func logUpgradableInstances(si []ccapi.ServiceInstance) {

}
