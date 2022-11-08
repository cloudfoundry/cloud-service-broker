package local

import (
	"log"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
)

func DeleteServiceKey(serviceName, keyName, cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	serviceInstance := lookupServiceInstanceByGUID(nameToID(serviceName))

	broker, err := testdrive.StartBroker(os.Args[0], pakDir, databasePath(), testdrive.WithOutputs(os.Stdout, os.Stderr))
	if err != nil {
		log.Fatal(err)
	}
	defer broker.Stop()

	if err := broker.DeleteBinding(serviceInstance, nameToID(keyName)); err != nil {
		log.Fatal(err)
	}
}
