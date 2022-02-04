// Package manifest is the data model for a manifest file.
package manifest

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/platform"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/validation"
)

type Manifest struct {
	PackVersion        int                 `yaml:"packversion"`
	Name               string              `yaml:"name"`
	Version            string              `yaml:"version"`
	Metadata           map[string]string   `yaml:"metadata"`
	Platforms          []platform.Platform `yaml:"platforms"`
	TerraformResources []TerraformResource `yaml:"terraform_binaries"`
	ServiceDefinitions []string            `yaml:"service_definitions"`
	Parameters         []Parameter         `yaml:"parameters"`
	RequiredEnvVars    []string            `yaml:"required_env_variables"`
	EnvConfigMapping   map[string]string   `yaml:"env_config_mapping"`
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
	defaults := 0
	for _, resource := range m.TerraformResources {
		if resource.Name == "terraform" {
			count++
			if resource.Default {
				defaults++
			}

			errs = errs.Also(validation.ErrIfDuplicate(resource.Version, "version", cache))
		}
	}

	switch {
	case count > 1 && defaults == 0:
		errs = errs.Also(&validation.FieldError{
			Message: "multiple Terraform versions, but none marked as default",
			Paths:   []string{"terraform_binaries"},
		})
	case count > 1 && defaults > 1:
		errs = errs.Also(&validation.FieldError{
			Message: "multiple Terraform versions, and multiple marked as default",
			Paths:   []string{"terraform_binaries"},
		})
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
