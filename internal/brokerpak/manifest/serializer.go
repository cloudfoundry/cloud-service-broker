package manifest

import "gopkg.in/yaml.v3"

func (m *Manifest) Serialize() ([]byte, error) {
	p := parser{
		PackVersion:        m.PackVersion,
		Name:               m.Name,
		Version:            m.Version,
		Metadata:           m.Metadata,
		Platforms:          m.Platforms,
		ServiceDefinitions: m.ServiceDefinitions,
		Parameters:         m.Parameters,
		RequiredEnvVars:    m.RequiredEnvVars,
		EnvConfigMapping:   m.EnvConfigMapping,
	}

	for _, v := range m.TerraformUpgradePath {
		p.TerraformUpgradePath = append(p.TerraformUpgradePath, TerraformUpgradePath{Version: v.String()})
	}

	for _, v := range m.TerraformVersions {
		p.TerraformResources = append(p.TerraformResources, TerraformResource{
			Name:        "terraform",
			Version:     v.Version.String(),
			Source:      v.Source,
			URLTemplate: v.URLTemplate,
			Default:     v.Default,
		})
	}
	for _, v := range m.TerraformProviders {
		p.TerraformResources = append(p.TerraformResources, TerraformResource{
			Name:        v.Name,
			Version:     v.Version.String(),
			Source:      v.Source,
			Provider:    v.Provider.String(),
			URLTemplate: v.URLTemplate,
		})
	}
	for _, v := range m.Binaries {
		p.TerraformResources = append(p.TerraformResources, TerraformResource{
			Name:        v.Name,
			Version:     v.Version,
			Source:      v.Source,
			URLTemplate: v.URLTemplate,
		})
	}

	return yaml.Marshal(p)
}
