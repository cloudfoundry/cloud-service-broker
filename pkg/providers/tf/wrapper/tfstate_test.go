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

import "fmt"

func ExampleNewTfstate_good() {
	state := `{
    "version": 4,
    "terraform_version": "0.12.20",
    "serial": 2,
    "outputs": {
        "hostname": {
          "value": "brokertemplate.instance.hostname",
          "type": "string"
        }
    },
    "resources": [
        {
          "module": "module.instance",
          "mode": "managed",
          "type": "google_sql_database",
          "name": "database",
          "provider": "provider.google",
          "instances": []
        },
        {
          "module": "module.instance",
          "mode": "managed",
          "type": "google_sql_database_instance",
          "name": "instance",
          "provider": "provider.google",
          "instances": []
        }
    ]
  }`

	_, err := NewTfstate([]byte(state))
	fmt.Printf("%v", err)

	// Output: <nil>
}

func ExampleNewTfstate_badVersion() {
	state := `{
    "version": 5,
    "terraform_version": "0.12.20",
    "serial": 2,
    "outputs": {
        "hostname": {
          "value": "brokertemplate.instance.hostname",
          "type": "string"
        }
    },
    "resources": [
        {
          "module": "module.instance",
          "mode": "managed",
          "type": "google_sql_database",
          "name": "database",
          "provider": "provider.google",
          "instances": []
        }
    ]
  }`

	_, err := NewTfstate([]byte(state))
	fmt.Printf("%v", err)

	// Output: unsupported tfstate version: 5
}

func ExampleTfstate_GetOutputs() {
	state := `{
    "version": 4,
    "terraform_version": "0.12.20",
    "serial": 2,
    "outputs": {
        "hostname": {
          "value": "somehost",
          "type": "string"
        }
    },
    "resources": [
        {
          "module": "module.instance",
          "mode": "managed",
          "type": "google_sql_database",
          "name": "database",
          "provider": "provider.google",
          "instances": []
        }
    ]
  }`

	tfstate, _ := NewTfstate([]byte(state))
	fmt.Printf("%v\n", tfstate.GetOutputs())

	// Output: map[hostname:somehost]
}
