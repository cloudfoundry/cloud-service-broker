package local

import (
	"log"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/client"
)

func RunExamples(serviceOfferingName, exampleName, cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	broker := startBroker(pakDir)
	defer broker.Stop()

	examples, err := listExamples(broker.Port)
	if err != nil {
		log.Fatal(err)
	}

	const jobCount = 1_000_000
	client.RunExamplesForService(examples, broker.Client, serviceOfferingName, exampleName, jobCount)
}
