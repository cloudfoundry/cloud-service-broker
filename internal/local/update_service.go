package local

import (
	"log"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
)

func UpdateService(name, plan, params, cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	serviceInstance := lookupServiceInstanceByGUID(nameToID(name))

	broker, err := testdrive.StartBroker(os.Args[0], pakDir, databasePath(), testdrive.WithOutputs(os.Stdout, os.Stderr))
	if err != nil {
		log.Fatal(err)
	}
	defer broker.Stop()

	opts := []testdrive.UpdateOption{testdrive.WithUpdateParams(params)}
	if plan != "" {
		planID := lookupPlanIDByName(broker.Client, serviceInstance.ServiceOfferingGUID, plan)
		opts = append(opts, testdrive.WithUpdatePlan(planID))
	}

	if err = broker.UpdateService(serviceInstance, opts...); err != nil {
		log.Fatal(err)
	}
}
