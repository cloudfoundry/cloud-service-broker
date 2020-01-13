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
	"testing"

	"github.com/pivotal/cloud-service-broker/pkg/validation"
)

func TestNewExampleManifest(t *testing.T) {
	exampleManifest := NewExampleManifest()

	if err := exampleManifest.Validate(); err != nil {
		t.Fatalf("example manifest should be valid, but got error: %v", err)
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
