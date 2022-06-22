package upgrader

import (
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/workers"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CCAPI
type CCAPI interface {
	GetServiceInstances([]string) ([]ccapi.ServiceInstance, error)
	GetServicePlans(string) ([]ccapi.Plan, error)
	UpgradeServiceInstance(string, string) error
	PollServiceInstance(string) (bool, error)
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

	serviceInstances, err := api.GetServiceInstances(planGUIDS)
	if err != nil {
		return err
	}

	type upgradeTask struct {
		Guid      string
		MIVersion string
	}

	upgradeQueue := make(chan upgradeTask)

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

	workers.Run(batchSize, func() {
		for instance := range upgradeQueue {
			api.UpgradeServiceInstance(instance.Guid, instance.MIVersion)
		}
	})

	return nil
}
