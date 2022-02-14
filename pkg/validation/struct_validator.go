// Copyright 2018 the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validation

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"

	"github.com/hashicorp/hcl/v2/hclparse"
)

var (
	osbNameRegex                = regexp.MustCompile(`^[a-zA-Z0-9-\.]+$`)
	terraformIdentifierRegex    = regexp.MustCompile(`^[a-z_]*$`)
	terraformAttributePathRegex = regexp.MustCompile(`^([-a-zA-Z0-9_-]*\.[-a-zA-Z0-9_-]*){2}`)
	jsonSchemaTypeRegex         = regexp.MustCompile(`^(|object|boolean|array|number|string|integer)$`)
	uuidRegex                   = regexp.MustCompile(`^[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}$`)
)

// ErrIfNotHCL returns an error if the value is not valid HCL.
func ErrIfNotHCL(value, field string) *FieldError {
	parser := hclparse.NewParser()
	if _, err := parser.ParseHCL([]byte(value), ""); err == nil {
		return nil
	}

	return &FieldError{
		Message: "invalid HCL",
		Paths:   []string{field},
	}
}

// ErrIfNotJSON returns an error if the value is not valid JSON.
func ErrIfNotJSON(value json.RawMessage, field string) *FieldError {
	if json.Valid(value) {
		return nil
	}

	return &FieldError{
		Message: "invalid JSON",
		Paths:   []string{field},
	}
}

// ErrIfBlank returns an error if the value is a blank string.
func ErrIfBlank(value, field string) *FieldError {
	if value == "" {
		return ErrMissingField(field)
	}

	return nil
}

// ErrIfNil returns an error if the value is nil.
func ErrIfNil(value interface{}, field string) *FieldError {
	if value == nil {
		return ErrMissingField(field)
	}

	return nil
}

// ErrIfNotOSBName returns an error if the value is not a valid OSB name.
func ErrIfNotOSBName(value, field string) *FieldError {
	return ErrIfNotMatch(value, osbNameRegex, field)
}

// ErrIfNotJSONSchemaType returns an error if the value is not a valid JSON
// schema type.
func ErrIfNotJSONSchemaType(value, field string) *FieldError {
	return ErrIfNotMatch(value, jsonSchemaTypeRegex, field)
}

// ErrIfNotTerraformAttributePath returns an error if the value is not a valid
// Terraform identifier for an attribute in the HCL.
func ErrIfNotTerraformAttributePath(value, field string) *FieldError {
	return ErrIfNotMatch(value, terraformAttributePathRegex, field)
}

// ErrIfNotTerraformIdentifier returns an error if the value is not a valid
// Terraform identifier.
func ErrIfNotTerraformIdentifier(value, field string) *FieldError {
	return ErrIfNotMatch(value, terraformIdentifierRegex, field)
}

// ErrIfNotUUID returns an error if the value is not a valid UUID.
func ErrIfNotUUID(value, field string) *FieldError {
	if uuidRegex.MatchString(value) {
		return nil
	}

	return &FieldError{
		Message: "field must be a UUID",
		Paths:   []string{field},
	}
}

// ErrIfNotURL returns an error if the value is not a valid URL.
func ErrIfNotURL(value, field string) *FieldError {
	// Validaiton inspired by: github.com/go-playground/validator/baked_in.go
	url, err := url.ParseRequestURI(value)
	if err != nil || url.Scheme == "" {
		return &FieldError{
			Message: "field must be a URL",
			Paths:   []string{field},
		}
	}

	return nil
}

// ErrIfNotMatch returns an error if the value doesn't match the regex.
func ErrIfNotMatch(value string, regex *regexp.Regexp, field string) *FieldError {
	if regex.MatchString(value) {
		return nil
	}

	return ErrMustMatch(value, regex, field)
}

// ErrMustMatch notifies the user a field must match a regex.
func ErrMustMatch(value string, regex *regexp.Regexp, field string) *FieldError {
	return &FieldError{
		Message: fmt.Sprintf("field must match '%s'", regex.String()),
		Paths:   []string{field},
	}
}

// ErrIfDuplicate returns error when a value is duplicated when it should be unique.
// State is stored in the cache which must be provided for every call in the set.
func ErrIfDuplicate(value, field string, cache map[string]struct{}) *FieldError {
	if _, dup := cache[value]; dup {
		return ErrDuplicate(value, field)
	}
	cache[value] = struct{}{}
	return nil
}

// ErrIfOutsideLength returns an error if the length of the specified string is outside
// of the specified length range
func ErrIfOutsideLength(value, field string, min, max int) *FieldError {
	l := len(value)
	if l > max || l < min {
		return ErrOutsideLength(l, min, max, field)
	}
	return nil
}

// Validatable indicates that a particular type may have its fields validated.
type Validatable interface {
	// Validate checks the validity of this types fields.
	Validate() *FieldError
}

// ValidatableTest is a standard way of testing Validatable types.
type ValidatableTest struct {
	Object Validatable
	Expect error
}

// Testable is a type derived from testing.T
type Testable interface {
	Errorf(format string, a ...interface{})
}

// Assert runs the validate function and fails Testable.
func (vt *ValidatableTest) Assert(t Testable) {
	actual := vt.Object.Validate()
	expect := vt.Expect

	switch {
	case expect == nil && actual == nil:
		// success
	case expect == nil && actual != nil:
		t.Errorf("expected: <nil> got: %s", actual.Error())
	case expect != nil && actual == nil:
		t.Errorf("expected: %s got: <nil>", expect.Error())
	case expect.Error() != actual.Error():
		t.Errorf("expected: %s got: %s", expect.Error(), actual.Error())
	}
}
