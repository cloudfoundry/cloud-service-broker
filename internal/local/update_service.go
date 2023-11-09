package local

import (
	"fmt"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	"github.com/pivotal-cf/brokerapi/v10/domain"
	"log"
)

func UpdateService(name, plan, params, cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	serviceInstance := lookupServiceInstanceByGUID(nameToID(name))
	deployment, _ := store().GetTerraformDeployment(fmt.Sprintf("tf:%s:", serviceInstance.GUID))
	tfVersion, _ := deployment.TFWorkspace().StateTFVersion()
	broker := startBroker(pakDir)
	defer broker.Stop()

	opts := []testdrive.UpdateOption{
		testdrive.WithUpdateParams(params),
		testdrive.WithUpdatePreviousValues(domain.PreviousValues{MaintenanceInfo: &domain.MaintenanceInfo{Version: tfVersion.String()}}),
	}

	if plan != "" {
		planID := lookupPlanIDByName(broker.Client, serviceInstance.ServiceOfferingGUID, plan)
		opts = append(opts, testdrive.WithUpdatePlan(planID))
	}

	if err := broker.UpdateService(serviceInstance, opts...); err != nil {
		log.Fatal(err)
	}
}
