package scheduler

import "github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"

type UpgradeTask struct {
	serviceInstanceGUID string
	newVersion          string
}

func ScheduleUpgrades(upgradeQueue chan UpgradeTask, instances []ccapi.ServiceInstance, planVersions map[string]string) {
	for _, instance := range instances {
		if instance.UpgradeAvailable {
			newVersion := planVersions[instance.Relationships.ServicePlan.Data.GUID]
			upgradeQueue <- UpgradeTask{
				serviceInstanceGUID: instance.GUID,
				newVersion:          newVersion,
			}
		}
	}
	close(upgradeQueue)
}
