package local

import (
	"log"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/testdrive"
)

func startBroker(pakDir string) *testdrive.Broker {
	broker, err := testdrive.StartBroker(os.Args[0], pakDir, databasePath(), testdrive.WithOutputs(os.Stdout, os.Stderr), testdrive.WithAllowedEnvs(passThroughEnvs()))
	if err != nil {
		log.Fatal(err)
	}

	return broker
}
