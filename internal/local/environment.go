package local

import (
	"log"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry/cloud-service-broker/pkg/brokerpak"
)

func manifestAllowedEnvs() (result []string) {
	data, err := os.ReadFile(brokerpak.ManifestName)
	if err != nil {
		log.Fatalf("could not read manifest file %q: %s", brokerpak.ManifestName, err)
	}

	m, err := manifest.Parse(data)
	if err != nil {
		log.Fatalf("could not parse manifest data: %s", err)
	}

	for k := range m.EnvConfigMapping {
		result = append(result, k)
	}

	return result
}
