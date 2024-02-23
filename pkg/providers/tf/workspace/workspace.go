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

// Package workspace is a model for Terraform workspaces
package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/command"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"

	"github.com/hashicorp/go-version"
)

// DefaultInstanceName is the default name of an instance of a particular module.
const (
	DefaultInstanceName = "instance"
)

// NewWorkspace creates a new TerraformWorkspace from a given template and variables to populate an instance of it.
// The created instance will have the name specified by the DefaultInstanceName constant.
func NewWorkspace(templateVars map[string]any,
	terraformTemplate string,
	terraformTemplates map[string]string,
	importParameterMappings []ParameterMapping,
	parametersToRemove []string,
	parametersToAdd []ParameterMapping) (*TerraformWorkspace, error) {
	tfModule := ModuleDefinition{
		Name:        "brokertemplate",
		Definition:  terraformTemplate,
		Definitions: terraformTemplates,
	}

	inputList, err := tfModule.Inputs()
	if err != nil {
		return nil, err
	}

	limitedConfig := make(map[string]any)
	for _, name := range inputList {
		limitedConfig[name] = templateVars[name]
	}

	workspace := TerraformWorkspace{
		Modules: []ModuleDefinition{tfModule},
		Instances: []ModuleInstance{
			{
				ModuleName:    tfModule.Name,
				InstanceName:  DefaultInstanceName,
				Configuration: limitedConfig,
			},
		},
		Transformer: TfTransformer{
			ParameterMappings:  importParameterMappings,
			ParametersToRemove: parametersToRemove,
			ParametersToAdd:    parametersToAdd,
		},
	}

	return &workspace, nil
}

// DeserializeWorkspace creates a new TerraformWorkspace from a given JSON
// serialization of one.
func DeserializeWorkspace(definition []byte) (*TerraformWorkspace, error) {
	ws := TerraformWorkspace{}
	if err := json.Unmarshal(definition, &ws); err != nil {
		return nil, err
	}

	return &ws, nil
}

// TerraformWorkspace represents the directory layout of a Terraform execution.
// The structure is strict, consisting of several Terraform modules and instances
// of those modules. The strictness is artificial, but maintains a clear
// separation between data and code.
//
// It manages the directory structure needed for the commands, serializing and
// deserializing Terraform state, and all the flags necessary to call Terraform.
//
// All public functions that shell out to Terraform maintain the following invariants:
// - The function blocks if another terraform shell is running.
// - The function updates the tfstate once finished.
// - The function creates and destroys its own dir.
type TerraformWorkspace struct {
	Modules   []ModuleDefinition `json:"modules"`
	Instances []ModuleInstance   `json:"instances"`
	State     []byte             `json:"tfstate"`

	Transformer TfTransformer `json:"transform"`

	dirLock sync.Mutex
	dir     string
}

func (workspace *TerraformWorkspace) StateTFVersion() (*version.Version, error) {
	if !workspace.HasState() {
		return nil, CannotReadVersionError{message: "workspace state not generated"}
	}

	var receiver struct {
		Version string `json:"terraform_version"`
	}
	if err := json.Unmarshal(workspace.State, &receiver); err != nil {
		return nil, CannotReadVersionError{message: fmt.Sprintf("invalid workspace state %s", err)}
	}

	return version.NewVersion(receiver.Version)
}

func (workspace *TerraformWorkspace) HasState() bool {
	return workspace.State != nil
}

// String returns a human-friendly representation of the workspace suitable for
// printing to the console.
func (workspace *TerraformWorkspace) String() string {
	var b strings.Builder

	b.WriteString("# Terraform Workspace\n")
	_, _ = fmt.Fprintf(&b, "modules: %d\n", len(workspace.Modules))
	_, _ = fmt.Fprintf(&b, "instances: %d\n", len(workspace.Instances))
	_, _ = fmt.Fprintln(&b)

	for _, instance := range workspace.Instances {
		_, _ = fmt.Fprintf(&b, "## Instance %q\n", instance.InstanceName)
		_, _ = fmt.Fprintf(&b, "module = %q\n", instance.ModuleName)

		for k, v := range instance.Configuration {
			_, _ = fmt.Fprintf(&b, "input.%s = %#v\n", k, v)
		}

		if outputs, err := workspace.Outputs(instance.InstanceName); err != nil {
			for k, v := range outputs {
				_, _ = fmt.Fprintf(&b, "output.%s = %#v\n", k, v)
			}
		}
		_, _ = fmt.Fprintln(&b)
	}

	return b.String()
}

// Serialize converts the TerraformWorkspace into a JSON string.
func (workspace *TerraformWorkspace) Serialize() (string, error) {
	ws, err := json.Marshal(workspace)
	if err != nil {
		return "", err
	}

	return string(ws), nil
}

// initializeFsFlat initializes simple terraform directory structure
func (workspace *TerraformWorkspace) initializeFsFlat() error {
	if len(workspace.Modules) != 1 {
		return fmt.Errorf("cannot build flat terraform workspace with multiple modules")
	}
	if len(workspace.Instances) != 1 {
		return fmt.Errorf("cannot build flat terraform workspace with multiple instances")
	}

	for name, tf := range workspace.Modules[0].Definitions {
		if err := os.WriteFile(path.Join(workspace.dir, fmt.Sprintf("%s.tf", name)), []byte(tf), 0755); err != nil {
			return err
		}
	}

	variables, err := json.MarshalIndent(workspace.Instances[0].Configuration, "", "  ")

	if err == nil {
		err = os.WriteFile(path.Join(workspace.dir, "terraform.tfvars.json"), variables, 0755)
	}
	return err
}

// initializeFsModules initializes multi-module terraform directory structure
func (workspace *TerraformWorkspace) initializeFsModules() error {
	outputs := make(map[string][]string)

	// write the modulesTerraformWorkspace
	for _, module := range workspace.Modules {
		parent := path.Join(workspace.dir, module.Name)
		if err := os.Mkdir(parent, 0755); err != nil {
			return err
		}

		if len(module.Definition) > 0 {
			if err := os.WriteFile(path.Join(parent, "definition.tf"), []byte(module.Definition), 0755); err != nil {
				return err
			}
		}

		for name, tf := range module.Definitions {
			if err := os.WriteFile(path.Join(parent, fmt.Sprintf("%s.tf", name)), []byte(tf), 0755); err != nil {
				return err
			}
		}

		var err error
		if outputs[module.Name], err = module.Outputs(); err != nil {
			return err
		}
	}

	// write the instances
	for _, instance := range workspace.Instances {
		output := outputs[instance.ModuleName]
		contents, err := instance.MarshalDefinition(output)
		if err != nil {
			return err
		}

		if err := os.WriteFile(path.Join(workspace.dir, instance.InstanceName+".tf.json"), contents, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (workspace *TerraformWorkspace) initializeFsWithoutTerraformInit() error {
	workspace.dirLock.Lock()
	// create a temp directory
	if dir, err := os.MkdirTemp("", "gsb"); err == nil {
		workspace.dir = dir
	} else {
		return err
	}

	var err error

	terraformLen := 0
	for _, module := range workspace.Modules {
		terraformLen += len(module.Definition)
		for _, def := range module.Definitions {
			terraformLen += len(def)
		}
	}

	if len(workspace.Modules) == 1 && len(workspace.Modules[0].Definition) == 0 && terraformLen > 0 {
		err = workspace.initializeFsFlat()
	} else {
		err = workspace.initializeFsModules()
	}

	if err != nil {
		return err
	}

	// write the state if it exists
	if len(workspace.State) > 0 {
		if err = os.WriteFile(workspace.tfStatePath(), workspace.State, 0755); err != nil {
			return err
		}
	}
	return nil
}

// TeardownFs removes the directory we executed Terraform in and updates the
// state from it.
func (workspace *TerraformWorkspace) teardownFs() error {
	bytes, err := os.ReadFile(workspace.tfStatePath())
	if err != nil {
		return err
	}

	workspace.State = bytes

	if err := os.RemoveAll(workspace.dir); err != nil {
		return err
	}

	workspace.dir = ""
	workspace.dirLock.Unlock()
	return nil
}

// Outputs gets the Terraform outputs from the state for the instance with the
// given name. This function DOES NOT invoke Terraform and instead uses the stored state.
// If no instance exists with the given name, it could be that Terraform pruned it due
// to having no contents so a blank map is returned.
func (workspace *TerraformWorkspace) Outputs(_ string) (map[string]any, error) {
	state, err := NewTfstate(workspace.State)
	if err != nil {
		return nil, fmt.Errorf("error creating TF state: %w", err)
	}

	// All root project modules get put under the "root" namespace
	return state.GetOutputs(), nil
}

func (workspace *TerraformWorkspace) Execute(ctx context.Context, terraformExecutor executor.TerraformExecutor, commands ...command.TerraformCommand) (executor.ExecutionOutput, error) {
	err := workspace.initializeFsWithoutTerraformInit()
	defer func(w *TerraformWorkspace) {
		_ = w.teardownFs()
	}(workspace)
	if err != nil {
		return executor.ExecutionOutput{}, err
	}
	var lastExecutionOutput executor.ExecutionOutput

	for _, cmd := range commands {
		c := exec.Command("terraform", cmd.Command()...)
		c.Env = os.Environ()
		c.Dir = workspace.dir

		lastExecutionOutput, err = terraformExecutor.Execute(ctx, c)
		if err != nil {
			return executor.ExecutionOutput{}, err
		}
	}
	return lastExecutionOutput, nil
}

func (workspace *TerraformWorkspace) tfStatePath() string {
	return path.Join(workspace.dir, "terraform.tfstate")
}

func (workspace *TerraformWorkspace) UpdateInstanceConfiguration(templateVars map[string]any) error {
	// we may be doing this twice in the case of dynamic HCL, that is fine.
	inputList, err := workspace.Modules[0].Inputs()
	if err != nil {
		return err
	}
	limitedConfig := make(map[string]any)
	for _, name := range inputList {
		limitedConfig[name] = templateVars[name]
	}
	workspace.Instances[0].Configuration = limitedConfig
	return nil
}

func (workspace *TerraformWorkspace) ModuleDefinitions() []ModuleDefinition {
	return workspace.Modules
}

func (workspace *TerraformWorkspace) ModuleInstances() []ModuleInstance {
	return workspace.Instances
}
