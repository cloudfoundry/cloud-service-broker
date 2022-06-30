package upgrader

import (
	"fmt"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/workers"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . CFClient
type CFClient interface {
	GetServiceInstances([]string) ([]ccapi.ServiceInstance, error)
	GetServicePlans(string) ([]ccapi.Plan, error)
	UpgradeServiceInstance(string, string) error
}

//counterfeiter:generate . Logger
type Logger interface {
	Printf(format string, a ...any)
	UpgradeStarting(guid string)
	UpgradeSucceeded(guid string, duration time.Duration)
	UpgradeFailed(guid string, duration time.Duration, err error)
	InitialTotals(totalServiceInstances, totalUpgradableServiceInstances int)
	FinalTotals()
}

func Upgrade(api CFClient, brokerName string, batchSize int, log Logger) error {
	plans, err := api.GetServicePlans(brokerName)
	if err != nil {
		return err
	}

	if len(plans) == 0 {
		return fmt.Errorf(fmt.Sprintf("no service plans available for broker: %s", brokerName))
	}

	var planGUIDS []string
	planVersions := make(map[string]string)

	for _, plan := range plans {
		planGUIDS = append(planGUIDS, plan.GUID)
		planVersions[plan.GUID] = plan.MaintenanceInfoVersion
	}

	log.Printf("discovering service instances for broker: %s", brokerName)

	serviceInstances, err := api.GetServiceInstances(planGUIDS)
	if err != nil {
		return err
	}

	var upgradableInstances []ccapi.ServiceInstance
	for _, i := range serviceInstances {
		if i.UpgradeAvailable {
			upgradableInstances = append(upgradableInstances, i)
		}
	}
	if len(upgradableInstances) == 0 {
		log.Printf("no instances available to upgrade")
		return nil
	}

	log.InitialTotals(len(serviceInstances), len(upgradableInstances))

	type upgradeTask struct {
		ServiceInstanceGUID    string
		MaintenanceInfoVersion string
	}

	upgradeQueue := make(chan upgradeTask)
	go func() {
		for _, instance := range upgradableInstances {
			upgradeQueue <- upgradeTask{
				ServiceInstanceGUID:    instance.GUID,
				MaintenanceInfoVersion: planVersions[instance.PlanGUID],
			}
		}
		close(upgradeQueue)
	}()

	workers.Run(batchSize, func() {
		for instance := range upgradeQueue {
			start := time.Now()
			log.UpgradeStarting(instance.ServiceInstanceGUID)
			err := api.UpgradeServiceInstance(instance.ServiceInstanceGUID, instance.MaintenanceInfoVersion)
			switch err {
			case nil:
				log.UpgradeSucceeded(instance.ServiceInstanceGUID, time.Since(start))
			default:
				log.UpgradeFailed(instance.ServiceInstanceGUID, time.Since(start), err)
			}
		}
	})

	log.FinalTotals()
	return nil
}
