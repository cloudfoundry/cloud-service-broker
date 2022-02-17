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
	"fmt"
	"sort"

	"github.com/cloudfoundry/cloud-service-broker/pkg/validation"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// ModuleDefinition represents a module in a Terraform workspace.
type ModuleDefinition struct {
	Name        string
	Definition  string
	Definitions map[string]string
}

var _ (validation.Validatable) = (*ModuleDefinition)(nil)

// Validate checks the validity of the ModuleDefinition struct.
func (module *ModuleDefinition) Validate() (errs *validation.FieldError) {
	for name, definition := range module.Definitions {
		errs = errs.Also(validation.ErrIfNotHCL(definition, fmt.Sprintf("Definitions[%v]", name)))
	}
	return errs.Also(
		validation.ErrIfBlank(module.Name, "Name"),
		validation.ErrIfNotTerraformIdentifier(module.Name, "Name"),
		validation.ErrIfNotHCL(module.Definition, "Definition"),
	)
}

func decode(body string) (hcl.Blocks, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCL([]byte(body), "")
	if diags.HasErrors() {
		return hcl.Blocks{}, fmt.Errorf(diags.Error())
	}
	schema := hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "variable",
				LabelNames: []string{"type"},
			},
			{
				Type:       "output",
				LabelNames: []string{"value"},
			},
		},
	}
	content, _, diags := f.Body.PartialContent(&schema)
	if diags.HasErrors() {
		return hcl.Blocks{}, fmt.Errorf(diags.Error())
	}

	return content.Blocks, nil
}

func (module *ModuleDefinition) decode() (blocks hcl.Blocks, err error) {
	blocks, err = decode(module.Definition)

	if err == nil {
		for name, definition := range module.Definitions {
			newBlocks, err := decode(definition)
			if err != nil {
				return blocks, fmt.Errorf("%v decoding definitions[%v]", err, name)
			}
			blocks = append(blocks, newBlocks...)
		}
	}

	return
}

// Inputs gets the input parameter names for the module.
func (module *ModuleDefinition) Inputs() ([]string, error) {
	blocks, err := module.decode()

	return sortedKeys(blocks.OfType("variable")), err
}

// Outputs gets the output parameter names for the module.
func (module *ModuleDefinition) Outputs() ([]string, error) {
	blocks, err := module.decode()

	return sortedKeys(blocks.OfType("output")), err
}

func sortedKeys(m hcl.Blocks) []string {
	var keys []string
	for _, block := range m {
		keys = append(keys, block.Labels...)
	}

	sort.Slice(keys, func(i int, j int) bool { return keys[i] < keys[j] })
	return keys
}
