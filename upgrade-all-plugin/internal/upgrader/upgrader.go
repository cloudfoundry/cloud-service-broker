package upgrader

import (
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/workers"
)

type Plan struct {
	GUID                   string
	MaintenanceInfoVersion string
}

type ServiceInstance struct {
	GUID             string
	UpgradeAvailable bool
	PlanGUID         string
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CCAPI
type CCAPI interface {
	GetServiceInstances([]string) ([]ServiceInstance, error)
	GetServicePlans(string) ([]Plan, error)
	UpgradeServiceInstance(string) error
	PollServiceInstance(string) (bool, error)
}

func Upgrade(api CCAPI, brokerName string) error {
	plans, err := api.GetServicePlans(brokerName)
	if err != nil {
		return err
	}

	var planGUIDS []string
	for _, plan := range plans {
		planGUIDS = append(planGUIDS, plan.GUID)
	}

	serviceInstances, err := api.GetServiceInstances(planGUIDS)
	if err != nil {
		return err
	}

	upgradeQueue := make(chan string)

	go func() {
		for _, instance := range serviceInstances {
			if instance.UpgradeAvailable {
				upgradeQueue <- instance.GUID
			}
		}
		close(upgradeQueue)
	}()

	workers.Run(5, func() {
		for instance := range upgradeQueue {
			api.UpgradeServiceInstance(instance)
		}
	})

	return nil
}
