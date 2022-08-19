// Package manifest is the data model for a manifest file.
package manifest

import (
	"github.com/hashicorp/go-version"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"
	"github.com/cloudfoundry/cloud-service-broker/internal/tfproviderfqn"
)

// Manifest is the internal model for the brokerpak manifest
type Manifest struct {
	PackVersion                        int
	Name                               string
	Version                            string
	Metadata                           map[string]string
	Platforms                          []platform.Platform
	TerraformVersions                  []TerraformVersion
	TerraformProviders                 []TerraformProvider
	Binaries                           []Binary
	ServiceDefinitions                 []string
	Parameters                         []Parameter
	RequiredEnvVars                    []string
	EnvConfigMapping                   map[string]string
	TerraformUpgradePath               []*version.Version
	TerraformStateProviderReplacements map[string]string
}

type TerraformVersion struct {
	Version     *version.Version
	Default     bool
	Source      string
	URLTemplate string
}

type TerraformProvider struct {
	Name        string
	Version     *version.Version
	Source      string
	Provider    tfproviderfqn.TfProviderFQN
	URLTemplate string
}

type Binary struct {
	Name        string
	Version     string
	Source      string
	URLTemplate string
}
