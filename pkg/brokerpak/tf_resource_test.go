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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/validation"
)

func TestTerraformResource_Validate(t *testing.T) {
	cases := map[string]validation.ValidatableTest{
		"blank obj": {
			Object: &TerraformResource{},
			Expect: errors.New("missing field(s): name, source, version"),
		},
		"good obj": {
			Object: &TerraformResource{
				Name:    "foo",
				Version: "1.0",
				Source:  "github.com/myproject",
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			tc.Assert(t)
		})
	}
}

func TestTerraformResource_Url(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("Unable to get current working dir %v", err)
	}
	cases := map[string]struct {
		Resource	TerraformResource
		Plat        Platform
		ExpectedURL string
	}{
		"default": {
			Resource: TerraformResource{
				Name: "foo",
				Version: "1.0",
				Source: "github.com/myproject",
			},
			Plat: Platform{
				Os: "my_os",
				Arch: "my_arch",
			},
			ExpectedURL: fmt.Sprintf("https://releases.hashicorp.com/%s/%s/%s_%s_%s_%s.zip","foo", "1.0", "foo", "1.0", "my_os", "my_arch"),
		},
		"custom": {
			Resource: TerraformResource{
				Name: "foo",
				Version: "1.0",
				Source: "github.com/myproject",
				UrlTemplate: "https://myproject/${name}_${version}_${os}_${arch}",
			},
			Plat: Platform{
				Os: "my_os",
				Arch: "my_arch",
			},
			ExpectedURL: fmt.Sprintf("https://myproject/%s_%s_%s_%s","foo", "1.0", "my_os", "my_arch"),
		},
		"handles_relative_path": {
			Resource: TerraformResource{
				Name: "foo",
				Version: "1.0",
				Source: "github.com/myproject",
				UrlTemplate: "../test_path",
			},
			Plat: Platform{
				Os: "my_os",
				Arch: "my_arch",
			},
			ExpectedURL: fmt.Sprintf("%s/test_path", filepath.Dir(wd)),			
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			url := tc.Resource.Url(tc.Plat)
			if url != tc.ExpectedURL {
				t.Errorf("Expected URL to be %v, got %v", tc.ExpectedURL, url)
			}
		})
	}
}
