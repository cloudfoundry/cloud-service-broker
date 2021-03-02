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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils"
)

// DefaultInstanceName is the default name of an instance of a particular module.
const (
	DefaultInstanceName = "instance"
)

var (
	FsInitializationErr = errors.New("Filesystem must first be initialized.")
)

// ExecutionOutput captures output from tf cli execution
type ExecutionOutput struct {
	StdOut string
	StdErr string
}

// TerraformExecutor is the function that shells out to Terraform.
// It can intercept, modify or retry the given command.
type TerraformExecutor func(*exec.Cmd) (ExecutionOutput, error)

// NewWorkspace creates a new TerraformWorkspace from a given template and variables to populate an instance of it.
// The created instance will have the name specified by the DefaultInstanceName constant.
func NewWorkspace(templateVars map[string]interface{},
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

	limitedConfig := make(map[string]interface{})
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
func DeserializeWorkspace(definition string) (*TerraformWorkspace, error) {
	ws := TerraformWorkspace{}
	if err := json.Unmarshal([]byte(definition), &ws); err != nil {
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

	// Executor is a function that gets invoked to shell out to Terraform.
	// If left nil, the default executor is used.
	Executor    TerraformExecutor `json:"-"`
	Transformer TfTransformer     `json:"transform"`

	dirLock sync.Mutex
	dir     string
}

// String returns a human-friendly representation of the workspace suitable for
// printing to the console.
func (workspace *TerraformWorkspace) String() string {
	var b strings.Builder

	b.WriteString("# Terraform Workspace\n")
	fmt.Fprintf(&b, "modules: %d\n", len(workspace.Modules))
	fmt.Fprintf(&b, "instances: %d\n", len(workspace.Instances))
	fmt.Fprintln(&b)

	for _, instance := range workspace.Instances {
		fmt.Fprintf(&b, "## Instance %q\n", instance.InstanceName)
		fmt.Fprintf(&b, "module = %q\n", instance.ModuleName)

		for k, v := range instance.Configuration {
			fmt.Fprintf(&b, "input.%s = %#v\n", k, v)
		}

		if outputs, err := workspace.Outputs(instance.InstanceName); err != nil {
			for k, v := range outputs {
				fmt.Fprintf(&b, "output.%s = %#v\n", k, v)
			}
		}
		fmt.Fprintln(&b)
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

// initializedFsFlat initializes simple terraform directory structure
func (workspace *TerraformWorkspace) initializedFsFlat() error {
	if len(workspace.Modules) != 1 {
		return fmt.Errorf("Cannot build flat terraform workspace with multiple modules")
	}
	if len(workspace.Instances) != 1 {
		return fmt.Errorf("Cannot build flat terraform workspace with multiple instances")
	}

	for name, tf := range workspace.Modules[0].Definitions {
		if err := ioutil.WriteFile(path.Join(workspace.dir, fmt.Sprintf("%s.tf", name)), []byte(tf), 0755); err != nil {
			return err
		}
	}

	variables, err := json.MarshalIndent(workspace.Instances[0].Configuration, "", "  ")

	if err == nil {
		err = ioutil.WriteFile(path.Join(workspace.dir, "terraform.tfvars.json"), variables, 0755)
	}
	return err
}

// initializeFsModules initializes multimodule terrafrom directory structure
func (workspace *TerraformWorkspace) initializeFsModules() error {
	outputs := make(map[string][]string)

	// write the modulesTerraformWorkspace
	for _, module := range workspace.Modules {
		parent := path.Join(workspace.dir, module.Name)
		if err := os.Mkdir(parent, 0755); err != nil {
			return err
		}

		if len(module.Definition) > 0 {
			if err := ioutil.WriteFile(path.Join(parent, "definition.tf"), []byte(module.Definition), 0755); err != nil {
				return err
			}
		}

		for name, tf := range module.Definitions {
			if err := ioutil.WriteFile(path.Join(parent, fmt.Sprintf("%s.tf", name)), []byte(tf), 0755); err != nil {
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

		if err := ioutil.WriteFile(path.Join(workspace.dir, instance.InstanceName+".tf.json"), contents, 0755); err != nil {
			return err
		}
	}
	return nil
}

// initializeFs initializes the filesystem directory necessary to run Terraform.
func (workspace *TerraformWorkspace) initializeFs() error {
	workspace.dirLock.Lock()
	// create a temp directory
	if dir, err := ioutil.TempDir("", "gsb"); err == nil {
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
		err = workspace.initializedFsFlat()
	} else {
		err = workspace.initializeFsModules()
	}

	if err != nil {
		return err
	}

	// write the state if it exists
	if len(workspace.State) > 0 {
		if err := ioutil.WriteFile(workspace.tfStatePath(), workspace.State, 0755); err != nil {
			return err
		}
	}

	// run "terraform init"
	if _, err := workspace.runTf("init", "-no-color"); err != nil {
		return err
	}

	return nil
}

// TeardownFs removes the directory we executed Terraform in and updates the
// state from it.
func (workspace *TerraformWorkspace) teardownFs() error {
	bytes, err := ioutil.ReadFile(workspace.tfStatePath())
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
func (workspace *TerraformWorkspace) Outputs(instance string) (map[string]interface{}, error) {
	state, err := NewTfstate(workspace.State)
	if err != nil {
		return nil, err
	}

	// All root project modules get put under the "root" namespace
	return state.GetOutputs(), nil
}

// Validate runs `terraform Validate` on this workspace.
// This function blocks if another Terraform command is running on this workspace.
func (workspace *TerraformWorkspace) Validate() error {
	err := workspace.initializeFs()
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	_, err = workspace.runTf("validate", "-no-color")

	return err
}

// Apply runs `terraform apply` on this workspace.
// This function blocks if another Terraform command is running on this workspace.
func (workspace *TerraformWorkspace) Apply() error {
	err := workspace.initializeFs()
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	_, err = workspace.runTf("apply", "-auto-approve", "-no-color")
	return err
}

// Destroy runs `terraform destroy` on this workspace.
// This function blocks if another Terraform command is running on this workspace.
func (workspace *TerraformWorkspace) Destroy() error {
	err := workspace.initializeFs()
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	_, err = workspace.runTf("destroy", "-auto-approve", "-no-color")
	return err
}

// Apply runs `terraform import` on this workspace.
// This function blocks if another Terraform command is running on this workspace.
func (workspace *TerraformWorkspace) Import(resources map[string]string) error {
	err := workspace.initializeFs()
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	for resource, id := range resources {
		_, err = workspace.runTf("import", resource, id)
		if err != nil {
			return err
		}
	}

	return nil
}

// Apply runs `terraform show` on this workspace.
// This function blocks if another Terraform command is running on this workspace.
func (workspace *TerraformWorkspace) Show() (string, error) {
	err := workspace.initializeFs()
	defer workspace.teardownFs()
	if err != nil {
		return "", err
	}

	output, err := workspace.runTf("show", "-no-color")
	if err != nil {
		return "", err
	}

	return output.StdOut, nil
}

func (workspace *TerraformWorkspace) tfStatePath() string {
	return path.Join(workspace.dir, "terraform.tfstate")
}

func (workspace *TerraformWorkspace) runTf(subCommand string, args ...string) (ExecutionOutput, error) {
	sub := []string{subCommand}
	sub = append(sub, args...)

	c := exec.Command("terraform", sub...)
	c.Env = os.Environ()
	c.Dir = workspace.dir

	executor := DefaultExecutor
	if workspace.Executor != nil {
		executor = workspace.Executor
	}

	return executor(c)
}

// CustomEnvironmentExecutor sets custom environment variables on the Terraform
// execution.
func CustomEnvironmentExecutor(environment map[string]string, wrapped TerraformExecutor) TerraformExecutor {
	return func(c *exec.Cmd) (ExecutionOutput, error) {
		for k, v := range environment {
			c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
		}

		return wrapped(c)
	}
}

func updatePath(vars []string, path string) string {
	for _, envVar := range vars {
		varPair := strings.Split(envVar, "=")
		if strings.TrimSpace(varPair[0]) == "PATH" && len(varPair) > 1 {
			return fmt.Sprintf("PATH=%s:%s", path, strings.TrimSpace(varPair[1]))
		}
	}
	return fmt.Sprintf("PATH=%s", path)
}

// CustomTerraformExecutor executes a custom Terraform binary that uses plugins
// from a given plugin directory rather than the Terraform that's on the PATH
// which will download provider binaries from the web.
func CustomTerraformExecutor(tfBinaryPath, tfPluginDir string, wrapped TerraformExecutor) TerraformExecutor {
	return func(c *exec.Cmd) (ExecutionOutput, error) {

		// Add the -get-plugins=false and -plugin-dir={tfPluginDir} after the
		// sub-command to force Terraform to use a particular plugin.
		subCommand := c.Args[1]
		subCommandArgs := c.Args[2:]

		if subCommand == "init" {
			subCommandArgs = append([]string{"-get-plugins=false", fmt.Sprintf("-plugin-dir=%s", tfPluginDir)}, subCommandArgs...)
		}

		allArgs := append([]string{subCommand}, subCommandArgs...)
		newCmd := exec.Command(tfBinaryPath, allArgs...)
		newCmd.Dir = c.Dir
		newCmd.Env = append(c.Env, updatePath(c.Env, tfPluginDir))
		return wrapped(newCmd)
	}
}

// DefaultExecutor is the default executor that shells out to Terraform
// and logs results to stdout.
func DefaultExecutor(c *exec.Cmd) (ExecutionOutput, error) {
	logger := utils.NewLogger("terraform@" + c.Dir)

	logger.Info("starting process", lager.Data{
		"path": c.Path,
		"args": c.Args,
		"dir":  c.Dir,
	})

	stderr, err := c.StderrPipe()
	if err != nil {
		return ExecutionOutput{}, fmt.Errorf("Failed to get stderr pipe for terraform execution: %v", err)
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		return ExecutionOutput{}, fmt.Errorf("Failed to get stdout pipe for terraform execution: %v", err)
	}

	if err := c.Start(); err != nil {
		return ExecutionOutput{}, fmt.Errorf("Failed to execute terraform: %v", err)
	}

	output, _ := ioutil.ReadAll(stdout)
	errors, _ := ioutil.ReadAll(stderr)

	err = c.Wait()

	if err != nil ||
		len(errors) > 0 {
		logger.Error("terraform execution failed", err, lager.Data{
			"errors": string(errors),
		})
	}

	logger.Debug("terraform output", lager.Data{
		"output": string(output),
	})

	if err != nil {
		return ExecutionOutput{}, fmt.Errorf("%s %v", strings.ReplaceAll(string(errors), "\n", ""), err)
	}

	return ExecutionOutput{
		StdErr: string(errors),
		StdOut: string(output),
	}, nil
}
