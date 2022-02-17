package manifest

import "github.com/cloudfoundry/cloud-service-broker/pkg/validation"

type Parameter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

var _ validation.Validatable = (*Parameter)(nil)

// Validate implements validation.Validatable.
func (param *Parameter) Validate() (errs *validation.FieldError) {
	return errs.Also(
		validation.ErrIfBlank(param.Name, "name"),
		validation.ErrIfBlank(param.Description, "description"),
	)
}
