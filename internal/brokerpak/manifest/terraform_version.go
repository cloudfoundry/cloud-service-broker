package manifest

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

func (m *Manifest) DefaultTerraformVersion() (*version.Version, error) {
	var versions []string
	for _, r := range m.TerraformResources {
		if r.Name == "terraform" {
			if r.Default {
				return version.NewVersion(r.Version)
			}
			versions = append(versions, r.Version)
		}
	}

	switch len(versions) {
	case 0:
		return &version.Version{}, fmt.Errorf("terraform not found")
	case 1:
		return version.NewVersion(versions[0])
	default:
		return &version.Version{}, fmt.Errorf("no default terraform found")
	}
}
