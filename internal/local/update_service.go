package local

import (
	"fmt"
	"log"

	"github.com/pivotal-cf/brokerapi/v11/domain"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/testdrive"
)

func UpdateService(name, plan, params, cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	serviceInstance := lookupServiceInstanceByGUID(nameToID(name))
	deployment, err := store().GetTerraformDeployment(fmt.Sprintf("tf:%s:", serviceInstance.GUID))
	if err != nil {
		log.Fatal(err)
	}
	tfVersion, err := deployment.TFWorkspace().StateTFVersion()
	if err != nil {
		log.Fatal(err)
	}
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
