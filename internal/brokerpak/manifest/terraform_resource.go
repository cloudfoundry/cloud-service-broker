package manifest

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cloud-service-broker/pkg/validation"
)

type TerraformResource struct {
	// Name holds the name of this resource. e.g. terraform-provider-google-beta
	Name string `yaml:"name"`

	// Version holds the version of the resource e.g. 1.19.0
	Version string `yaml:"version"`

	// Source holds the URI of an archive that contains the source code for this release.
	Source string `yaml:"source"`

	// Provider holds path to extract the provider path.
	Provider string `yaml:"provider,omitempty"`

	// URLTemplate holds a custom URL template to get the release of the given tool.
	// Parameters available are ${name}, ${version}, ${os}, and ${arch}.
	// If non is specified HashicorpUrlTemplate is used.
	URLTemplate string `yaml:"url_template,omitempty"`

	// Default is used to mark the default Terraform version when there is more than one
	Default bool `yaml:"default"`
}

func (tr TerraformResource) GetProviderNamespace() string {
	if tr.Provider != "" {
		return strings.Split(tr.Provider, "/")[0]
	}

	return "hashicorp"
}

func (tr TerraformResource) GetProviderType() string {
	if tr.Provider != "" {
		return strings.Split(tr.Provider, "/")[1]
	}

	parts := strings.SplitAfterN(tr.Name, "terraform-provider-", 2)
	if len(parts) == 2 {
		return parts[1]
	}

	return tr.Name
}

var _ validation.Validatable = (*TerraformResource)(nil)

// Validate implements validation.Validatable.
func (tr *TerraformResource) Validate() (errs *validation.FieldError) {
	return errs.Also(
		validation.ErrIfBlank(tr.Name, "name"),
		validation.ErrIfBlank(tr.Version, "version"),
		tr.validateDefault(),
		tr.validateModule(),
	)
}

func (tr *TerraformResource) validateDefault() *validation.FieldError {
	if tr.Default && tr.Name != "terraform" {
		return &validation.FieldError{
			Message: "This field is only valid for `terraform`",
			Paths:   []string{"default"},
		}
	}

	return nil
}

func (tr *TerraformResource) validateModule() *validation.FieldError {
	if tr.Provider != "" {
		parts := strings.Split(tr.Provider, "/")
		if len(parts) == 0 || len(parts) > 2 {
			return &validation.FieldError{
				Message: fmt.Sprintf("This field is only valid for %s. Provide module as `namespace/name`.", tr.Name),
				Paths:   []string{"module"},
			}
		}
	}

	return nil
}
