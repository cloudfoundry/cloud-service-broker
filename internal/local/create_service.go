package local

import (
	"log"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/testdrive"
)

func CreateService(service, plan, name, params, cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	broker := startBroker(pakDir)
	defer broker.Stop()

	serviceID, planID := lookupServiceIDsByName(broker.Client, service, plan)
	instanceID := nameToID(name)

	_, err := broker.Provision(serviceID, planID, testdrive.WithProvisionServiceInstanceGUID(instanceID), testdrive.WithProvisionParams(params))
	if err != nil {
		log.Fatal(err)
	}
}
