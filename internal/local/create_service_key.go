package local

import (
	"fmt"
	"log"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
)

func CreateServiceKey(serviceName, keyName, params, cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	serviceInstance := lookupServiceInstanceByGUID(nameToID(serviceName))

	broker, err := testdrive.StartBroker(os.Args[0], pakDir, databasePath(), testdrive.WithOutputs(os.Stdout, os.Stderr))
	if err != nil {
		log.Fatal(err)
	}
	defer broker.Stop()

	binding, err := broker.CreateBinding(serviceInstance, testdrive.WithBindingGUID(nameToID(keyName)), testdrive.WithBindingParams(params))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nService Key: %s\n", binding.Body)
}
