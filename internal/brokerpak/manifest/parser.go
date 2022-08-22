package manifest

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/go-version"
	"gopkg.in/yaml.v3"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"
	"github.com/cloudfoundry/cloud-service-broker/internal/tfproviderfqn"
	"github.com/cloudfoundry/cloud-service-broker/pkg/validation"
)

// parser is used to parse the brokerpak manifest.
// In code, we should use the Manifest type
type parser struct {
	PackVersion                        int                    `yaml:"packversion"`
	Name                               string                 `yaml:"name"`
	Version                            string                 `yaml:"version"`
	Metadata                           map[string]string      `yaml:"metadata"`
	Platforms                          []platform.Platform    `yaml:"platforms"`
	TerraformResources                 []TerraformResource    `yaml:"terraform_binaries"`
	ServiceDefinitions                 []string               `yaml:"service_definitions"`
	Parameters                         []Parameter            `yaml:"parameters"`
	RequiredEnvVars                    []string               `yaml:"required_env_variables"`
	EnvConfigMapping                   map[string]string      `yaml:"env_config_mapping"`
	TerraformUpgradePath               []TerraformUpgradePath `yaml:"terraform_upgrade_path,omitempty"`
	TerraformStateProviderReplacements map[string]string      `yaml:"terraform_state_provider_replacements,omitempty"`
}

func Parse(input []byte) (*Manifest, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(input))
	decoder.KnownFields(true)

	var receiver parser
	if err := decoder.Decode(&receiver); err != nil {
		return nil, fmt.Errorf("error parsing manifest: %w", err)
	}

	var result Manifest
	result.PackVersion = receiver.PackVersion
	result.Name = receiver.Name
	result.Version = receiver.Version
	result.Metadata = receiver.Metadata
	result.Platforms = receiver.Platforms
	result.ServiceDefinitions = receiver.ServiceDefinitions
	result.Parameters = receiver.Parameters
	result.RequiredEnvVars = receiver.RequiredEnvVars
	result.EnvConfigMapping = receiver.EnvConfigMapping
	result.TerraformStateProviderReplacements = receiver.TerraformStateProviderReplacements

	steps := []func() *validation.FieldError{
		func() (errs *validation.FieldError) {
			return receiver.Validate()
		},
		func() (errs *validation.FieldError) {
			result.TerraformUpgradePath, errs = parseTerraformUpgradePath(receiver)
			return
		},
		func() (errs *validation.FieldError) {
			result.TerraformVersions, result.TerraformProviders, result.Binaries, errs = parseTerraformResources(receiver)
			return
		},
	}

	var errs *validation.FieldError
	for _, p := range steps {
		errs = errs.Also(p())
	}

	if errs != nil {
		return nil, fmt.Errorf("error validating manifest: %w", errs)
	}

	return &result, nil
}

func parseTerraformUpgradePath(p parser) (result []*version.Version, errs *validation.FieldError) {
	availableTerraformVersions := make(map[string]bool)
	for _, v := range p.TerraformResources {
		if v.Name == "terraform" {
			availableTerraformVersions[v.Version] = v.Default
		}
	}

	cur := version.Must(version.NewVersion("0.0.0"))
	for i, v := range p.TerraformUpgradePath {
		isDefault, available := availableTerraformVersions[v.Version]
		ver, err := version.NewVersion(v.Version)
		switch {
		case err != nil:
			errs = errs.Also(validation.ErrInvalidValue(v.Version, "version")).ViaFieldIndex("terraform_upgrade_path", i)
		case !ver.GreaterThan(cur):
			errs = errs.Also((&validation.FieldError{
				Message: fmt.Sprintf("expect versions to be in ascending order: %q <= %q", v.Version, cur.String()),
				Paths:   []string{"version"},
			}).ViaFieldIndex("terraform_upgrade_path", i))
		case !available:
			errs = errs.Also((&validation.FieldError{
				Message: fmt.Sprintf("no corresponding terrafom resource for terraform version %q", v.Version),
				Paths:   []string{"version"},
			}).ViaFieldIndex("terraform_upgrade_path", i))
		case i == len(p.TerraformUpgradePath)-1 && !isDefault:
			errs = errs.Also((&validation.FieldError{
				Message: "upgrade path does not terminate at default version",
				Paths:   []string{"version"},
			}).ViaFieldIndex("terraform_upgrade_path", i))
		}

		cur = ver
		result = append(result, ver)
	}

	return result, errs
}

func parseTerraformResources(p parser) (versions []TerraformVersion, providers []TerraformProvider, binaries []Binary, errs *validation.FieldError) {
	if len(p.TerraformResources) == 0 {
		return nil, nil, nil, validation.ErrMissingField("terraform_binaries")
	}

	terraformVersionCache := make(map[string]struct{})
	for i, r := range p.TerraformResources {
		var (
			ver         *version.Version
			providerFQN tfproviderfqn.TfProviderFQN
			err         error
		)

		errs = errs.Also(
			validation.ErrIfBlank(r.Name, "name").ViaFieldIndex("terraform_binaries", i),
			validation.ErrIfBlank(r.Version, "version").ViaFieldIndex("terraform_binaries", i),
		)

		if r.resourceType() == terraformProvider {
			providerFQN, err = tfproviderfqn.New(r.Name, r.Provider)
			if err != nil {
				errs = errs.Also((&validation.FieldError{
					Message: err.Error(),
					Paths:   []string{"provider"},
				}).ViaFieldIndex("terraform_binaries", i))
			}
		}

		if r.Default && r.resourceType() != terraformVersion {
			errs = errs.Also((&validation.FieldError{
				Message: "This field is only valid for `terraform`",
				Paths:   []string{"default"},
			}).ViaFieldIndex("terraform_binaries", i))
		}

		ver, err = version.NewVersion(r.Version)
		if err != nil && (r.resourceType() == terraformVersion || r.resourceType() == terraformProvider) {
			errs = errs.Also((&validation.FieldError{
				Message: err.Error(),
				Paths:   []string{"version"},
			}).ViaFieldIndex("terraform_binaries", i))
		}

		if r.resourceType() == terraformVersion {
			errs = errs.Also(validation.ErrIfDuplicate(r.Version, "version", terraformVersionCache))
		}

		switch r.resourceType() {
		case terraformVersion:
			versions = append(versions, TerraformVersion{
				Version:     ver,
				Default:     r.Default,
				Source:      r.Source,
				URLTemplate: r.URLTemplate,
			})
		case terraformProvider:
			providers = append(providers, TerraformProvider{
				Name:        r.Name,
				Version:     ver,
				Source:      r.Source,
				Provider:    providerFQN,
				URLTemplate: r.URLTemplate,
			})
		case otherBinary:
			binaries = append(binaries, Binary{
				Name:        r.Name,
				Version:     r.Version,
				Source:      r.Source,
				URLTemplate: r.URLTemplate,
			})
		}
	}

	var defaultTerraformVersions []*version.Version
	highestTerraformVersion := version.Must(version.NewVersion("0.0.0"))
	for _, v := range versions {
		if v.Default {
			defaultTerraformVersions = append(defaultTerraformVersions, v.Version)
		}
		if v.Version != nil && v.Version.GreaterThan(highestTerraformVersion) {
			highestTerraformVersion = v.Version
		}
	}

	switch {
	case len(versions) > 1 && len(defaultTerraformVersions) == 0:
		errs = errs.Also(&validation.FieldError{
			Message: "multiple Terraform versions, but none marked as default",
			Paths:   []string{"terraform_binaries"},
		})
	case len(versions) > 1 && len(defaultTerraformVersions) > 1:
		errs = errs.Also(&validation.FieldError{
			Message: "multiple Terraform versions, and multiple marked as default",
			Paths:   []string{"terraform_binaries"},
		})
	case len(defaultTerraformVersions) == 1 && !defaultTerraformVersions[0].Equal(highestTerraformVersion):
		errs = errs.Also(&validation.FieldError{
			Message: "default version of Terraform must be the highest version",
			Paths:   []string{"terraform_binaries"},
		})
	}

	return versions, providers, binaries, errs
}

var _ validation.Validatable = (*parser)(nil)

func (m *parser) Validate() (errs *validation.FieldError) {
	validators := []func() *validation.FieldError{
		m.validatePackageVersion,
		m.validateName,
		m.validateVersion,
		m.validatePlatforms,
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

func (m *parser) validatePackageVersion() *validation.FieldError {
	switch m.PackVersion {
	case 0:
		return validation.ErrMissingField("packversion")
	case 1:
		return nil
	default:
		return validation.ErrInvalidValue(m.PackVersion, "packversion")
	}
}

func (m *parser) validateName() *validation.FieldError {
	return validation.ErrIfBlank(m.Name, "name")
}

func (m *parser) validateVersion() *validation.FieldError {
	return validation.ErrIfBlank(m.Version, "version")
}

func (m *parser) validatePlatforms() *validation.FieldError {
	if len(m.Platforms) == 0 {
		return validation.ErrMissingField("platforms")
	}

	var errs *validation.FieldError
	for i, p := range m.Platforms {
		errs = errs.Also(p.Validate().ViaFieldIndex("platforms", i))
	}

	return errs
}

func (m *parser) validateServiceDefinitions() *validation.FieldError {
	if len(m.ServiceDefinitions) == 0 {
		return validation.ErrMissingField("service_definitions")
	}

	return nil
}

func (m *parser) validateParameters() (errs *validation.FieldError) {
	for i, parameter := range m.Parameters {
		errs = errs.Also(parameter.Validate().ViaFieldIndex("parameters", i))
	}

	return errs
}
