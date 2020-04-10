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

package brokerpak

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-yaml/yaml"
	"github.com/pivotal/cloud-service-broker/pkg/validation"
)

func TestNewExampleManifest(t *testing.T) {
	exampleManifest := NewExampleManifest()

	if err := exampleManifest.Validate(); err != nil {
		t.Fatalf("example manifest should be valid, but got error: %v", err)
	}
}

func TestUnmarshalManifest(t *testing.T) {
	cases := map[string]struct {
		yaml     string
		expected Manifest
	}{
		"normal": {
			yaml: `packversion: 1
name: my-services-pack
version: 0.1.0
metadata:
  author: VMware
platforms:
- os: linux
  arch: amd64
terraform_binaries:
- name: terraform
  version: 0.12.23
  source: https://github.com/hashicorp/terraform/archive/v0.12.23.zip  
required_env_variables:
- ARM_SUBSCRIPTION_ID
service_definitions:
- example-service-definition.yml
env_config_mapping:
  ARM_SUBSCRIPTION_ID: azure.subscription_id`,
			expected: Manifest{
				PackVersion: 1,
				Name:        "my-services-pack",
				Version:     "0.1.0",
				Metadata: map[string]string{
					"author": "VMware",
				},
				Platforms: []Platform{
					{Os: "linux", Arch: "amd64"},
				},
				TerraformResources: []TerraformResource{
					{
						Name:    "terraform",
						Version: "0.12.23",
						Source:  "https://github.com/hashicorp/terraform/archive/v0.12.23.zip",
					},
				},
				ServiceDefinitions: []string{"example-service-definition.yml"},
				RequiredEnvVars:    []string{"ARM_SUBSCRIPTION_ID"},
				EnvConfigMapping:   map[string]string{"ARM_SUBSCRIPTION_ID": "azure.subscription_id"},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			var actual Manifest
			err := yaml.Unmarshal([]byte(tc.yaml), &actual)
			if err != nil {
				t.Fatalf("failed to unmarshal yaml manifest: %v", err)
			}
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Fatalf("Expected: %v Actual: %v", tc.expected, actual)
			}
		})
	}
}

func TestManifestParameter_Validate(t *testing.T) {
	cases := map[string]validation.ValidatableTest{
		"blank obj": {
			Object: &ManifestParameter{},
			Expect: errors.New("missing field(s): description, name"),
		},
		"good obj": {
			Object: &ManifestParameter{
				Name:        "TEST",
				Description: "Usage goes here",
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			tc.Assert(t)
		})
	}
}
