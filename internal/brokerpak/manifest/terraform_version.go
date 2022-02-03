package manifest

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

func (m *Manifest) GetTerraformVersion() (*version.Version, error) {
	for _, r := range m.TerraformResources {
		if r.Name == "terraform" {
			return version.NewVersion(r.Version)
		}
	}
	return &version.Version{}, fmt.Errorf("terraform provider not found")
}
