package local

import (
	"log"

	"code.cloudfoundry.org/brokerapi/v13/domain"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/testdrive"
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
