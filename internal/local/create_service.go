package local

import (
	"log"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
)

func CreateService(service, plan, name, params, cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	broker, err := testdrive.StartBroker(os.Args[0], pakDir, databasePath(), testdrive.WithOutputs(os.Stdout, os.Stderr))
	if err != nil {
		log.Fatal(err)
	}
	defer broker.Stop()

	serviceID, planID := lookupServiceIDsByName(broker.Client, service, plan)
	instanceID := nameToID(name)

	_, err = broker.Provision(serviceID, planID, testdrive.WithProvisionServiceInstanceGUID(instanceID), testdrive.WithProvisionParams(params))
	if err != nil {
		log.Fatal(err)
	}
}
