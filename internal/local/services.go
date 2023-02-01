package local

import (
	"log"
)

func Services(cachePath string) {
	ids, err := store().GetServiceInstancesIDs()
	if err != nil {
		log.Fatal(err)
	}

	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	broker := startBroker(pakDir)
	defer broker.Stop()

	tp := newTablePrinter("Name", "Service offering", "Plan", "GUID")
	for _, id := range ids {
		instance := lookupServiceInstanceByGUID(id)
		serviceName, planName := lookupServiceNamesByID(broker.Client, instance.ServiceOfferingGUID, instance.ServicePlanGUID)
		tp.row(idToName(id), serviceName, planName, id)
	}
	tp.print()
}
