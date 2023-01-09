package local

import (
	"log"
)

func DeleteService(name, cachePath string) {
	instance := lookupServiceInstanceByGUID(nameToID(name))

	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	broker := startBroker(pakDir)
	defer broker.Stop()

	if err := broker.Deprovision(instance); err != nil {
		log.Fatal(err)
	}
}
