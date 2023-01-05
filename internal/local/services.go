package local

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.StripEscape)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Name\tService offering\tPlan")
	_, _ = fmt.Fprintln(w, "----\t----------------\t----")
	for _, id := range ids {
		instance := lookupServiceInstanceByGUID(id)
		serviceName, planName := lookupServiceNamesByID(broker.Client, instance.ServiceOfferingGUID, instance.ServicePlanGUID)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", idToName(id), serviceName, planName)
	}
	_, _ = fmt.Fprintln(w)
	_ = w.Flush()
}
