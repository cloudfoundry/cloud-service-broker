package local

import (
	"log"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
)

func DeleteService(name, cachePath string) {
	instance := lookupServiceInstanceByGUID(nameToID(name))

	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	broker, err := testdrive.StartBroker(os.Args[0], pakDir, databasePath(), testdrive.WithOutputs(os.Stdout, os.Stderr))
	if err != nil {
		log.Fatal(err)
	}
	defer broker.Stop()

	if err := broker.Deprovision(instance); err != nil {
		log.Fatal(err)
	}
}
