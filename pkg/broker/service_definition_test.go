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

package broker

import (
	"encoding/json"
	"testing"

	"github.com/pivotal-cf/brokerapi"
)

func TestServiceDefinition_CheckProhibitUpdate(t *testing.T) {
	svcDef := ServiceDefinition{
		ProvisionInputVariables: []BrokerVariable{
			{
				FieldName:      "prohibited",
				ProhibitUpdate: true,
			},
			{
				FieldName: "allowed",
			},
		},
	}

	cases := map[string]struct {
		rawParams      string
		expectedResult bool
		expectErr      bool
	}{
		"allowed": {
			rawParams:      `{"allowed":"some_val"}`,
			expectedResult: true,
			expectErr:      false,
		},
		"prohibited": {
			rawParams:      `{"prohibited":"some_val"}`,
			expectedResult: false,
			expectErr:      false,
		},
		"bad json": {
			rawParams:      `{"bogus"}`,
			expectedResult: false,
			expectErr:      true,
		},
		"empty": {
			expectedResult: true,
			expectErr:      false,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			actual, err := svcDef.AllowedUpdate(brokerapi.UpdateDetails{
				RawParameters: json.RawMessage(tc.rawParams),
			})
			if (err == nil) == tc.expectErr {
				t.Errorf("Did not get exected error result")

			}
			if actual != tc.expectedResult {
				t.Errorf("Expected result from allowed update check on %v to be: %v, got: %v", tc.rawParams, tc.expectedResult, actual)
			}
		})
	}
}
