package local

import (
	"log"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry/cloud-service-broker/pkg/brokerpak"
	"github.com/cloudfoundry/cloud-service-broker/pkg/featureflags"
)

func passThroughEnvs() []string {
	result := append(manifestAllowedEnvs(), "TF_LOG", "TF_LOG_CORE", "TF_LOG_PROVIDER", "TF_LOG_PATH")
	result = append(result, featureflags.AllFeatureFlagEnvVars...)
	return result
}

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
