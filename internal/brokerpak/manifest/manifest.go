// Package manifest is the data model for a manifest file.
package manifest

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"
	"github.com/cloudfoundry/cloud-service-broker/pkg/validation"
	"github.com/hashicorp/go-version"
)

type Manifest struct {
	PackVersion          int                    `yaml:"packversion"`
	Name                 string                 `yaml:"name"`
	Version              string                 `yaml:"version"`
	Metadata             map[string]string      `yaml:"metadata"`
	Platforms            []platform.Platform    `yaml:"platforms"`
	TerraformResources   []TerraformResource    `yaml:"terraform_binaries"`
	ServiceDefinitions   []string               `yaml:"service_definitions"`
	Parameters           []Parameter            `yaml:"parameters"`
	RequiredEnvVars      []string               `yaml:"required_env_variables"`
	EnvConfigMapping     map[string]string      `yaml:"env_config_mapping"`
	TerraformUpgradePath []TerraformUpgradePath `yaml:"terraform_upgrade_path"`
}

var _ validation.Validatable = (*Manifest)(nil)

func (m *Manifest) Validate() (errs *validation.FieldError) {
	validators := []func() *validation.FieldError{
		m.validatePackageVersion,
		m.validateName,
		m.validateVersion,
		m.validatePlatforms,
		m.validateTerraformResources,
		m.validateTerraforms,
		m.validateTerraformUpgradePath,
		m.validateServiceDefinitions,
		m.validateParameters,
	}

	for _, v := range validators {
		if err := v(); err != nil {
			errs = errs.Also(err)
		}
	}

	return errs
}

func (m *Manifest) validatePackageVersion() *validation.FieldError {
	switch m.PackVersion {
	case 0:
		return validation.ErrMissingField("packversion")
	case 1:
		return nil
	default:
		return validation.ErrInvalidValue(m.PackVersion, "packversion")
	}
}

func (m *Manifest) validateName() *validation.FieldError {
	return validation.ErrIfBlank(m.Name, "name")
}

func (m *Manifest) validateVersion() *validation.FieldError {
	return validation.ErrIfBlank(m.Version, "version")
}

func (m *Manifest) validatePlatforms() *validation.FieldError {
	if len(m.Platforms) == 0 {
		return validation.ErrMissingField("platforms")
	}

	var errs *validation.FieldError
	for i, platform := range m.Platforms {
		errs = errs.Also(platform.Validate().ViaFieldIndex("platforms", i))
	}

	return errs
}

func (m *Manifest) validateTerraformResources() *validation.FieldError {
	if len(m.TerraformResources) == 0 {
		return validation.ErrMissingField("terraform_binaries")
	}

	var errs *validation.FieldError
	for i, resource := range m.TerraformResources {
		errs = errs.Also(resource.Validate().ViaFieldIndex("terraform_binaries", i))
	}

	return errs
}

func (m *Manifest) validateTerraforms() (errs *validation.FieldError) {
	cache := make(map[string]struct{})
	count := 0
	highest := version.Must(version.NewVersion("0.0.0"))
	var defaults []*version.Version
	for _, resource := range m.TerraformResources {
		if resource.Name == "terraform" {
			count++

			ver, err := version.NewVersion(resource.Version)
			switch {
			case err != nil:
				errs = errs.Also(validation.ErrInvalidValue(resource.Version, "version"))
			case ver.GreaterThan(highest):
				highest = ver
			}

			if resource.Default {
				defaults = append(defaults, ver)
			}

			errs = errs.Also(validation.ErrIfDuplicate(resource.Version, "version", cache))
		}
	}

	switch {
	case count > 1 && len(defaults) == 0:
		errs = errs.Also(&validation.FieldError{
			Message: "multiple Terraform versions, but none marked as default",
			Paths:   []string{"terraform_binaries"},
		})
	case count > 1 && len(defaults) > 1:
		errs = errs.Also(&validation.FieldError{
			Message: "multiple Terraform versions, and multiple marked as default",
			Paths:   []string{"terraform_binaries"},
		})
	case len(defaults) == 1 && !defaults[0].Equal(highest):
		errs = errs.Also(&validation.FieldError{
			Message: "default version of Terraform must be the highest version",
			Paths:   []string{"terraform_binaries"},
		})
	}

	return errs
}

func (m *Manifest) validateTerraformUpgradePath() (errs *validation.FieldError) {
	available := make(map[string]bool)
	for _, v := range m.TerraformResources {
		if v.Name == "terraform" {
			available[v.Version] = true
		}
	}

	cur := version.Must(version.NewVersion("0.0.0"))
	for i, v := range m.TerraformUpgradePath {
		ver, err := version.NewVersion(v.Version)
		switch {
		case err != nil:
			errs = errs.Also(validation.ErrInvalidValue(v.Version, "version")).ViaFieldIndex("terraform_upgrade_path", i)
		case !ver.GreaterThan(cur):
			errs = errs.Also((&validation.FieldError{
				Message: fmt.Sprintf("expect versions to be in ascending order: %q <= %q", v.Version, cur.String()),
				Paths:   []string{"version"},
			}).ViaFieldIndex("terraform_upgrade_path", i))
		case !available[v.Version]:
			errs = errs.Also((&validation.FieldError{
				Message: fmt.Sprintf("no corresponding terrafom resource for terraform version %q", v.Version),
				Paths:   []string{"version"},
			}).ViaFieldIndex("terraform_upgrade_path", i))
		}

		cur = ver
	}

	return errs
}

func (m *Manifest) validateServiceDefinitions() *validation.FieldError {
	if len(m.ServiceDefinitions) == 0 {
		return validation.ErrMissingField("service_definitions")
	}

	return nil
}

func (m *Manifest) validateParameters() (errs *validation.FieldError) {
	for i, parameter := range m.Parameters {
		errs = errs.Also(parameter.Validate().ViaFieldIndex("parameters", i))
	}

	return errs
}
