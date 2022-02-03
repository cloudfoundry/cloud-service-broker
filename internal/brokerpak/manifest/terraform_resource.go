package manifest

import "github.com/cloudfoundry-incubator/cloud-service-broker/pkg/validation"

type TerraformResource struct {
	// Name holds the name of this resource. e.g. terraform-provider-google-beta
	Name string `yaml:"name"`

	// Version holds the version of the resource e.g. 1.19.0
	Version string `yaml:"version"`

	// Source holds the URI of an archive that contains the source code for this release.
	Source string `yaml:"source"`

	// UrlTemplate holds a custom URL template to get the release of the given tool.
	// Parameters available are ${name}, ${version}, ${os}, and ${arch}.
	// If non is specified HashicorpUrlTemplate is used.
	URLTemplate string `yaml:"url_template,omitempty"`
}

var _ validation.Validatable = (*TerraformResource)(nil)

// Validate implements validation.Validatable.
func (tr *TerraformResource) Validate() (errs *validation.FieldError) {
	return errs.Also(
		validation.ErrIfBlank(tr.Name, "name"),
		validation.ErrIfBlank(tr.Version, "version"),
	)
}
