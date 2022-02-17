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

package platform_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"
	"github.com/cloudfoundry/cloud-service-broker/pkg/validation"
)

func ExamplePlatform_String() {
	p := platform.Platform{Os: "bsd", Arch: "amd64"}
	fmt.Println(p.String())

	// Output: bsd/amd64
}

func ExamplePlatform_Equals() {
	p := platform.Platform{Os: "beos", Arch: "webasm"}
	fmt.Println(p.Equals(p))
	fmt.Println(p.Equals(platform.CurrentPlatform()))

	// Output: true
	// false
}

func ExamplePlatform_MatchesCurrent() {
	fmt.Println(platform.CurrentPlatform().MatchesCurrent())

	// Output: true
}

func TestPlatform_Validate(t *testing.T) {
	cases := map[string]validation.ValidatableTest{
		"blank obj": {
			Object: &platform.Platform{},
			Expect: errors.New("missing field(s): arch, os"),
		},
		"good obj": {
			Object: &platform.Platform{
				Os:   "linux",
				Arch: "amd64",
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			tc.Assert(t)
		})
	}
}
