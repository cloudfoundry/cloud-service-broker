package local

import (
	"log"
)

func DeleteServiceKey(serviceName, keyName, cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	serviceInstance := lookupServiceInstanceByGUID(nameToID(serviceName))

	broker := startBroker(pakDir)
	defer broker.Stop()

	if err := broker.DeleteBinding(serviceInstance, nameToID(keyName)); err != nil {
		log.Fatal(err)
	}
}
