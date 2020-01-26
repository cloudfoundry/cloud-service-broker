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
	"encoding/json"
	"fmt"
)

const (
	supportedTfStateVersion = 4
)

// NewTfstate deserializes a tfstate file.
func NewTfstate(stateFile []byte) (*Tfstate, error) {
	state := Tfstate{}
	if err := json.Unmarshal(stateFile, &state); err != nil {
		return nil, err
	}

	if state.Version != supportedTfStateVersion {
		return nil, fmt.Errorf("unsupported tfstate version: %d", state.Version)
	}

	return &state, nil
}

// Tfstate is a struct that can help us deserialize the tfstate JSON file.
type Tfstate struct {
	Version int             `json:"version"`
	Outputs map[string]struct {
		Type  string      `json:"type"`
		Value interface{} `json:"value"`
	} `json:"outputs"`
}

// GetOutputs gets the key/value outputs defined for a module.
func (module *Tfstate) GetOutputs() map[string]interface{} {
	out := make(map[string]interface{})

	for outputName, tfOutput := range module.Outputs {
		out[outputName] = tfOutput.Value
	}

	return out
}
