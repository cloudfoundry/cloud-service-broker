// Package passwordparser parses password data
package passwordparser

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/jsonry"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/validation"
)

type PasswordEntry struct {
	Label   string `jsonry:"label"`
	Secret  string `jsonry:"password.secret"`
	Primary bool   `jsonry:"primary"`
}

// UnmarshalJSON is implemented because JSONry doesn't currently support slices of structs
// so we unmarshal each struct individually
func (p *PasswordEntry) UnmarshalJSON(input []byte) error {
	return jsonry.Unmarshal(input, p)
}

func Parse(input any) ([]PasswordEntry, error) {
	result, err := parsePasswords(input)
	if err != nil {
		return nil, fmt.Errorf("password reading error: %w", err)
	}

	if err := validate(result); err != nil {
		return nil, fmt.Errorf("password configuration error: %w", err)
	}

	return result, nil
}

func parsePasswords(input any) ([]PasswordEntry, error) {
	switch v := input.(type) {
	case nil:
		return nil, nil
	case string:
		return parsePasswordsFromString(v)
	case []any:
		return parsePasswordsFromMapSlice(v)
	default:
		return nil, fmt.Errorf("password configuration type error, expected string or object array, got %T", input)
	}
}

func parsePasswordsFromString(input string) ([]PasswordEntry, error) {
	if len(input) == 0 {
		return nil, nil
	}

	var receiver []PasswordEntry
	if err := json.Unmarshal([]byte(input), &receiver); err != nil {
		return nil, fmt.Errorf("password configuration string could not be parsed as JSON: %w", err)
	}

	return receiver, nil
}

func parsePasswordsFromMapSlice(input []any) ([]PasswordEntry, error) {
	if len(input) == 0 {
		return nil, nil
	}

	asJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("error coding passwords as JSON: %w", err)
	}
	return parsePasswordsFromString(string(asJSON))
}

func validate(passwordEntries []PasswordEntry) (errs *validation.FieldError) {
	labels := make(map[string]struct{})
	primaries := 0
	for i, p := range passwordEntries {
		if p.Primary {
			primaries++
		}
		errs = errs.Also(
			validation.ErrIfOutsideLength(p.Secret, "secret.password", 20, 1024).ViaIndex(i),
			validation.ErrIfOutsideLength(p.Label, "label", 5, 20).ViaIndex(i),
			validation.ErrIfDuplicate(p.Label, "label", labels).ViaIndex(i),
		)
	}

	switch primaries {
	case 0, 1:
		return errs
	default:
		return errs.Also(&validation.FieldError{
			Message: "expected exactly one primary, got multiple; mark one password as primary and others as non-primary but do not remove them",
			Paths:   []string{"[].primary"},
		})
	}
}
