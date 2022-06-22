package upgrader

import (
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

func Upgrade(api CCAPI, brokerName string, batchSize int, l *log.Logger) error {
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

	l.Printf("Discovering service instances for broker: %s\n", brokerName)

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
	if len(upgradableInstances) == 0 {
		l.Printf("no instances available to upgrade\n")
		return nil
	}

	l.Printf("---\n"+
		"Total instances: %d\n"+
		"Upgradable instances: %d\n"+
		"---\n",
		len(serviceInstances),
		len(upgradableInstances))

	var upgraded int32
	upgradeComplete := make(chan bool)

	go logUpgradeProgress(upgradeComplete, &upgraded, len(upgradableInstances), l)

	l.Printf("Starting upgrade...\n")

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
				addFailedInstance(instance.Guid, err.Error())
				continue
			}
			atomic.AddInt32(&upgraded, 1)
		}
	})

	upgradeComplete <- true

	logUpgradeComplete(upgraded, failedInstances, l)

	return nil
}

func logUpgradeProgress(complete chan bool, upgraded *int32, upgradable int, l *log.Logger) {
	for {
		select {
		case <-complete:
			return
		case <-time.After(time.Minute):
			l.Printf("Upgraded %d/%d\n", *upgraded, upgradable)
		}
	}
}

func logUpgradeComplete(upgraded int32, failedInstances map[string]string, l *log.Logger) {
	l.Printf("---\n"+
		"Finished upgrade:\n"+
		"Total instances upgraded: %d\n",
		upgraded)

	if len(failedInstances) > 0 {
		l.Printf(
			"Failed to upgrade instances:\n" +
				"GUID\tError\n")

		for k, v := range failedInstances {
			l.Printf("%s\t%s\n", k, v)
		}
	}
}
