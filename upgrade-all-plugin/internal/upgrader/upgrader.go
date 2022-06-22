package upgrader

import (
	"fmt"
	"log"
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

func Upgrade(api CCAPI, brokerName string, batchSize int, log *log.Logger) error {
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

	log.Printf("Discovering service instances for broker: %s\n", brokerName)

	serviceInstances, err := api.GetServiceInstances(planGUIDS)
	if err != nil {
		return err
	}

	type upgradeTask struct {
		Guid      string
		MIVersion string
	}

	var upgradableInstances []ccapi.ServiceInstance
	for _, i := range serviceInstances {
		if i.UpgradeAvailable {
			upgradableInstances = append(upgradableInstances, i)
		}
	}

	log.Printf("---\n"+
		"Total instances: %d\n"+
		"Upgradable instances: %d\n"+
		"---\n",
		len(serviceInstances),
		len(upgradableInstances))

	var upgraded int32
	upgradeComplete := make(chan bool)

	go logUpgradeProgress(upgradeComplete, &upgraded, len(upgradableInstances))

	log.Printf("Starting upgrade...\n")

	upgradeQueue := make(chan upgradeTask)
	go func() {
		for _, instance := range upgradableInstances {
			upgradeQueue <- upgradeTask{
				Guid:      instance.GUID,
				MIVersion: planVersions[instance.PlanGUID],
			}
		}
		close(upgradeQueue)
	}()

	failedInstances := make(map[string]string)
	addFailedInstance := func(instance, description string) {
		var lock sync.Mutex
		lock.Lock()
		defer lock.Unlock()
		failedInstances[instance] = description
	}

	workers.Run(batchSize, func() {
		for instance := range upgradeQueue {
			err := api.UpgradeServiceInstance(instance.Guid, instance.MIVersion)
			if err != nil {
				log.Printf("error upgrading service instance: %s\n", err)
				addFailedInstance(instance.Guid, err.Error())
				continue
			}
			atomic.AddInt32(&upgraded, 1)
		}
	})

	upgradeComplete <- true

	logUpgradeComplete(upgraded, failedInstances)

	return nil
}

func logUpgradeProgress(complete chan bool, upgraded *int32, upgradable int) {
	for {
		select {
		case <-complete:
			return
		case <-time.After(time.Minute):
			fmt.Printf("Upgraded %d/%d\n", *upgraded, upgradable)
		}
	}
}

func logUpgradeComplete(upgraded int32, failedInstances map[string]string) {
	log.Printf("---\n"+
		"Finished upgrade:\n"+
		"Total instances upgraded: %d\n",
		upgraded)

	if len(failedInstances) > 0 {
		log.Printf(
			"Failed to upgrade instances:\n" +
				"GUID\tError\n")

		for k, v := range failedInstances {
			log.Printf("%s\t%s\n", k, v)
		}
	}
}
