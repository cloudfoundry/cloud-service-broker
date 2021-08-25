package encryption

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/validation"
)

type PasswordConfig struct {
	ID       string   `json:"guid"`
	Label    string   `json:"label"`
	Primary  bool     `json:"primary"`
	Password Password `json:"encryption_key"`
}

type Password struct {
	Secret string `json:"secret"`
}

type PasswordConfigs []PasswordConfig

func (passwords PasswordConfigs) Validate() error {
	if len(passwords) == 0 {
		return fmt.Errorf("no encryption keys were provided")
	}

	var errs *validation.FieldError
	primaryPasswords := 0
	IDs := make(map[string]struct{})
	labels := make(map[string]struct{})
	for i, password := range passwords {
		errs = errs.Also(
			password.Validate().ViaFieldIndex("Key", i),
			validation.ErrIfDuplicate("guid", password.ID, IDs).ViaFieldIndex("Key", i),
			validation.ErrIfDuplicate("label", password.Label, labels).ViaFieldIndex("Key", i),
		)
		if password.Primary {
			primaryPasswords++
		}
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}

	switch primaryPasswords {
	case 0:
		return fmt.Errorf("no encryption key is marked as primary")
	case 1:
		break
	default:
		return fmt.Errorf("several encryption keys are marked as primary")
	}

	return nil
}

func (password *PasswordConfig) Validate() (errs *validation.FieldError) {
	errs = errs.Also(
		validation.ErrIfNotUUID(password.ID, "guid"),
		validation.ErrIfBlank(password.Label, "label"),
		validation.ErrIfNotLength(password.Label, 5, 1024, "label"),
		validation.ErrIfBlank(password.Password.Secret, "encryption_key.secret"),
		validation.ErrIfNotLength(password.Password.Secret, 20, 1024, "encryption_key.secret"),
	)

	return errs
}
