package manifest

import "github.com/hashicorp/go-version"

type TerraformUpgradePath struct {
	Version string `yaml:"version"`
}

func (p TerraformUpgradePath) GetTerraformVersion() *version.Version {
	return version.Must(version.NewVersion(p.Version))
}
