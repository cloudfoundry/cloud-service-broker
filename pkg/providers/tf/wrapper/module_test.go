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

package wrapper

import (
	"strings"
	"testing"
)

func compareStringArrays(a1, a2 []string) bool {
	if len(a1) != len(a2) {
		return false
	}
	for n := range a1 {
		if a1[n] != a2[n] {
			return false
		}
	}
	return true
}

func TestModuleDefinition_Inputs(t *testing.T) {
	cases := map[string]struct {
		Module ModuleDefinition
		Inputs []string
	}{
		"definition": {
			Module: ModuleDefinition{
				Name: "cloud_storage",
				Definition: `
		    variable name {type = "string"}
		    variable storage_class {type = "string"}

		    resource "google_storage_bucket" "bucket" {
		      name     = "${var.name}"
		      storage_class = "${var.storage_class}"
		    }
		`,
			},
			Inputs: []string{"name", "storage_class"},
		},
		"definitions": {
			Module: ModuleDefinition{
				Name: "cloud_storage",
				Definition: `
			    resource "google_storage_bucket" "bucket" {
			      name     = "${var.name}"
			      storage_class = "${var.storage_class}"
			    }`,
				Definitions: map[string]string{"variables": `
				    variable name {type = "string"}
					variable storage_class {type = "string"}
				`},
			},
			Inputs: []string{"name", "storage_class"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			inputs, err := tc.Module.Inputs()
			if err != nil {
				t.Fatalf("Expected no error but got: %v", err)
			}
			if !compareStringArrays(inputs, tc.Inputs) {
				t.Fatalf("Expected to get inputs %v, but got %v", tc.Inputs, inputs)
			}
		})
	}
}

func TestModuleDefinition_Outputs(t *testing.T) {
	cases := map[string]struct {
		Module  ModuleDefinition
		Outputs []string
	}{
		"definition": {
			Module: ModuleDefinition{
				Name: "cloud_storage",
				Definition: `
		    resource "google_storage_bucket" "bucket" {
		      name     = "my-bucket"
		      storage_class = "STANDARD"
		    }

		    output id {value = "${google_storage_bucket.bucket.id}"}
			output bucket_name {value = "my-bucket"}
			`,
			},
			Outputs: []string{"bucket_name", "id"},
		},
		"definitions": {
			Module: ModuleDefinition{
				Name: "cloud_storage",
				Definition: `
		    resource "google_storage_bucket" "bucket" {
		      name     = "my-bucket"
		      storage_class = "STANDARD"
			}`,
				Definitions: map[string]string{"outputs": `
			    output id {value = "${google_storage_bucket.bucket.id}"}
				output bucket_name {value = "my-bucket"}
				`},
			},
			Outputs: []string{"bucket_name", "id"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			outputs, err := tc.Module.Outputs()
			if err != nil {
				t.Fatalf("Expected no error but got: %v", err)
			}
			if !compareStringArrays(outputs, tc.Outputs) {
				t.Fatalf("Expected to get ouputs %v, but got %v", tc.Outputs, outputs)
			}
		})
	}
}

func TestModuleDefinition_Validate(t *testing.T) {
	cases := map[string]struct {
		Module      ModuleDefinition
		ErrContains string
	}{
		"nominal": {
			Module: ModuleDefinition{
				Name: "my_module",
				Definition: `
          resource "google_storage_bucket" "bucket" {
            name     = "my-bucket"
            storage_class = "STANDARD"
          }`,
			},
			ErrContains: "",
		},
		"bad-name": {
			Module: ModuleDefinition{
				Name: "my module",
				Definition: `
          resource "google_storage_bucket" "bucket" {
            name     = "my-bucket"
            storage_class = "STANDARD"
          }`,
			},
			ErrContains: "field must match '^[a-z_]*$': Name",
		},
		"bad-hcl": {
			Module: ModuleDefinition{
				Name: "my_module",
				Definition: `
          resource "bucket" {
            name     = "my-bucket"`,
			},
			ErrContains: "invalid HCL: Definition",
		},
		"nominal-definitions": {
			Module: ModuleDefinition{
				Name: "my_module",
				Definitions: map[string]string{"main": `
          resource "google_storage_bucket" "bucket" {
            name     = "my-bucket"
            storage_class = "STANDARD"
          }`},
			},
			ErrContains: "",
		},
		"bad-hcl-definitions": {
			Module: ModuleDefinition{
				Name: "my_module",
				Definitions: map[string]string{"main": `
          resource "bucket" {
			name     = "my-bucket"`,
				},
			},
			ErrContains: "invalid HCL: Definitions[main]",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			err := tc.Module.Validate()
			if tc.ErrContains == "" {
				if err != nil {
					t.Fatalf("Expected no error but got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error containing %q but got nil", tc.ErrContains)
				}
				if !strings.Contains(err.Error(), tc.ErrContains) {
					t.Fatalf("Expected error containing %q but got %v", tc.ErrContains, err)
				}
			}
		})
	}
}
