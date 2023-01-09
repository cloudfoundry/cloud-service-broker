package local

import (
	"fmt"
	"log"

	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
)

func CreateServiceKey(serviceName, keyName, params, cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	serviceInstance := lookupServiceInstanceByGUID(nameToID(serviceName))

	broker := startBroker(pakDir)
	defer broker.Stop()

	binding, err := broker.CreateBinding(serviceInstance, testdrive.WithBindingGUID(nameToID(keyName)), testdrive.WithBindingParams(params))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nService Key: %s\n", binding.Body)
}
