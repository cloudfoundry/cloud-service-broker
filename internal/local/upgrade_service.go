package local

import (
	"log"

	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	"github.com/pivotal-cf/brokerapi/v9/domain"
)

func UpgradeService(name, previousVersion, cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	serviceInstance := lookupServiceInstanceByGUID(nameToID(name))
	planGUID := serviceInstance.ServicePlanGUID
	broker := startBroker(pakDir)
	defer broker.Stop()

	maintenanceInfo := lookupPlanMaintenanceInfoByGUID(broker.Client, serviceInstance.ServiceOfferingGUID, planGUID)

	previousMaintenanceInfo := domain.MaintenanceInfo{
		Version: previousVersion,
	}
	upgradeOption := testdrive.WithUpgradePreviousValues(domain.PreviousValues{MaintenanceInfo: &previousMaintenanceInfo, PlanID: planGUID, ServiceID: serviceInstance.ServiceOfferingGUID})
	if err := broker.UpgradeService(serviceInstance, maintenanceInfo.Version, upgradeOption); err != nil {
		log.Fatal(err)
	}
}
