package manifest

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

func (m *Manifest) DefaultTerraformVersion() (*version.Version, error) {
	var versions []*version.Version
	for _, r := range m.TerraformVersions {
		if r.Default {
			return r.Version, nil
		}
		versions = append(versions, r.Version)
	}

	switch len(versions) {
	case 0:
		return &version.Version{}, fmt.Errorf("terraform not found")
	case 1:
		return versions[0], nil
	default:
		return &version.Version{}, fmt.Errorf("no default terraform found")
	}
}
