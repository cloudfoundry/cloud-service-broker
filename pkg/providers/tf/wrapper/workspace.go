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
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/hashicorp/go-version"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/correlation"
)

// DefaultInstanceName is the default name of an instance of a particular module.
const (
	DefaultInstanceName = "instance"
)

// ExecutionOutput captures output from tf cli execution
type ExecutionOutput struct {
	StdOut string
	StdErr string
}

// TerraformExecutor is the function that shells out to Terraform.
// It can intercept, modify or retry the given command.
type TerraformExecutor func(context.Context, *exec.Cmd) (ExecutionOutput, error)

var planMessageMatcher = regexp.MustCompile(`Plan: \d+ to add, \d+ to change, (\d+) to destroy\.`)

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

// initializeFs initializes the filesystem directory necessary to run Terraform.
func (workspace *TerraformWorkspace) initializeFs(ctx context.Context) error {
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
		err = workspace.initializedFsFlat()
	} else {
		err = workspace.initializeFsModules()
	}

	if err != nil {
		return err
	}

	// write the state if it exists
	if len(workspace.State) > 0 {
		if err := os.WriteFile(workspace.tfStatePath(), workspace.State, 0755); err != nil {
			return err
		}
	}

	// run "terraform init"
	if _, err := workspace.runTf(ctx, "init", "-no-color"); err != nil {
		return err
	}

	return nil
}

func (workspace *TerraformWorkspace) AddSensitiveToOutputs() error {

	// For Azure
	//tfFilePoint, err := os.Open(path.Join(workspace.dir, workspace.Modules[0].Name, "definition.tf"))

	// For AWS
	tfFilePoint, err := os.Open(path.Join(workspace.dir, "outputs.tf"))
	if err != nil {
		return fmt.Errorf("unable to open tf file: %s", err)
	}
	defer tfFilePoint.Close()

	scanner := bufio.NewScanner(tfFilePoint)

	var tfString string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "# output") {
			continue
		}
		if strings.Contains(line, ")}") {
			line = strings.ReplaceAll(line, ")}", ")\n}")
		}
		if strings.Contains(line, "output") {
			line = strings.ReplaceAll(line, "{", "{\n ")
			line = strings.ReplaceAll(line, "}", "\n}")
		}
		tfString = tfString + "\n" + line
	}

	//For Azure
	//tfFile, diags := hclwrite.ParseConfig([]byte(tfString), workspace.Modules[0].Name, hcl.Pos{Line: 1, Column: 1})
	//For Aws
	tfFile, diags := hclwrite.ParseConfig([]byte(tfString), "outputs.tf", hcl.Pos{Line: 1, Column: 1})

	if diags.HasErrors() {
		return fmt.Errorf("errors: %s", diags)
	}

	for _, block := range tfFile.Body().Blocks() {
		if block.Type() == "output" {
			block.Body().SetAttributeValue("sensitive", cty.True)
		}
	}

	// For Azure
	//return os.WriteFile(path.Join(workspace.dir, workspace.Modules[0].Name, "definition.tf"), tfFile.Bytes(), 0755)
	//For AWS
	return os.WriteFile(path.Join(workspace.dir, "/outputs.tf"), tfFile.Bytes(), 0755)
}

func (workspace *TerraformWorkspace) initializeFsNoInit(ctx context.Context) error {
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
		err = workspace.initializedFsFlat()
	} else {
		err = workspace.initializeFsModules()
	}

	if err != nil {
		return err
	}

	// write the state if it exists
	if len(workspace.State) > 0 {
		if err := os.WriteFile(workspace.tfStatePath(), workspace.State, 0755); err != nil {
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
func (workspace *TerraformWorkspace) Outputs(instance string) (map[string]interface{}, error) {
	state, err := NewTfstate(workspace.State)
	if err != nil {
		return nil, fmt.Errorf("error creating TF state: %w", err)
	}

	// All root project modules get put under the "root" namespace
	return state.GetOutputs(), nil
}

// Validate runs `terraform Validate` on this workspace.
// This function blocks if another Terraform command is running on this workspace.
func (workspace *TerraformWorkspace) Validate(ctx context.Context) error {
	err := workspace.initializeFs(ctx)
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	_, err = workspace.runTf(ctx, "validate", "-no-color")

	return err
}

// Apply runs `terraform apply` on this workspace.
// This function blocks if another Terraform command is running on this workspace.
func (workspace *TerraformWorkspace) Apply(ctx context.Context) error {
	err := workspace.initializeFs(ctx)
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	_, err = workspace.runTf(ctx, "apply", "-auto-approve", "-no-color")
	return err
}

func (workspace *TerraformWorkspace) Plan(ctx context.Context) error {
	logger := utils.NewLogger("terraform-plan").WithData(correlation.ID(ctx))

	err := workspace.initializeFs(ctx)
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	output, err := workspace.runTf(ctx, "plan", "-no-color")
	if err != nil {
		return err
	}

	matches := planMessageMatcher.FindStringSubmatch(output.StdOut)
	switch {
	case len(matches) == 0: // presumably: "No changes. Infrastructure is up-to-date."
		logger.Info("no-match")
	case len(matches) == 2 && matches[1] == "0":
		logger.Info("no-destroyed")
	default:
		logger.Info("cancelling-destroy", lager.Data{"stdout": output.StdOut, "stderr": output.StdErr})
		return fmt.Errorf("terraform plan shows that resources would be destroyed - cancelling subsume")
	}

	return nil
}

// Destroy runs `terraform destroy` on this workspace.
// This function blocks if another Terraform command is running on this workspace.
func (workspace *TerraformWorkspace) Destroy(ctx context.Context) error {
	err := workspace.initializeFs(ctx)
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	_, err = workspace.runTf(ctx, "destroy", "-auto-approve", "-no-color")
	return err
}

// Import runs `terraform import` on this workspace.
// This function blocks if another Terraform command is running on this workspace.
func (workspace *TerraformWorkspace) Import(ctx context.Context, resources map[string]string) error {
	err := workspace.initializeFs(ctx)
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	for resource, id := range resources {
		_, err = workspace.runTf(ctx, "import", resource, id)
		if err != nil {
			return err
		}
	}

	return nil
}

// Show runs `terraform show` on this workspace.
// This function blocks if another Terraform command is running on this workspace.
func (workspace *TerraformWorkspace) Show(ctx context.Context) (string, error) {
	err := workspace.initializeFs(ctx)
	defer workspace.teardownFs()
	if err != nil {
		return "", err
	}

	output, err := workspace.runTf(ctx, "show", "-no-color")
	if err != nil {
		return "", err
	}

	return output.StdOut, nil
}

func (workspace *TerraformWorkspace) MigrateTo013(ctx context.Context) error {
	logger := utils.NewLogger("terraform-migrate-to-0.13").WithData(correlation.ID(ctx))

	err := workspace.initializeFsNoInit(ctx)
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	// need to check if flat file system or modules
	if _, err = os.Stat(workspace.dir + "/instance.tf.json"); errors.Is(err, os.ErrNotExist) {
		output, err := workspace.runTf(ctx, "0.13upgrade", "-yes", "-no-color")
		if err != nil {
			return err
		}
		logger.Info(output.StdOut)
	} else {
		output, err := workspace.runTf(ctx, "0.13upgrade", "-yes", "-no-color", "brokertemplate")
		if err != nil {
			return err
		}
		logger.Info(output.StdOut)
	}

	// state provider replace
	//Azure
	//providerMap := map[string]string{
	//	"registry.terraform.io/-/azurerm":    "registry.terraform.io/hashicorp/azurerm",
	//	"registry.terraform.io/-/random":     "registry.terraform.io/hashicorp/random",
	//	"registry.terraform.io/-/mysql":      "registry.terraform.io/hashicorp/mysql",
	//	"registry.terraform.io/-/null":       "registry.terraform.io/hashicorp/null",
	//	"registry.terraform.io/-/postgresql": "registry.terraform.io/hashicorp/postgresql",
	//}

	//AWS
	providerMap := map[string]string{
		"registry.terraform.io/-/aws":        "registry.terraform.io/hashicorp/aws",
		"registry.terraform.io/-/random":     "registry.terraform.io/hashicorp/random",
		"registry.terraform.io/-/mysql":      "registry.terraform.io/hashicorp/mysql",
		"registry.terraform.io/-/null":       "registry.terraform.io/hashicorp/null",
		"registry.terraform.io/-/postgresql": "registry.terraform.io/hashicorp/postgresql",
	}

	for oldProvider, newProvider := range providerMap {
		output, err := workspace.runTf(ctx, "state", "replace-provider", "-no-color", "-auto-approve", oldProvider, newProvider)
		if err != nil {
			return err
		}
		logger.Info(output.StdOut)
	}

	output, err := workspace.runTf(ctx, "init", "-no-color")
	if err != nil {
		return err
	}
	logger.Info(output.StdOut)

	output, err = workspace.runTf(ctx, "plan", "-no-color")
	if err != nil {
		return err
	}
	logger.Info(output.StdOut)

	output, err = workspace.runTf(ctx, "apply", "-auto-approve", "-no-color")
	if err != nil {
		return err
	}
	logger.Info(output.StdOut)

	return nil
}

func (workspace *TerraformWorkspace) MigrateTo014(ctx context.Context) error {
	logger := utils.NewLogger("terraform-migrate-to-0.14").WithData(correlation.ID(ctx))

	err := workspace.initializeFs(ctx)
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	output, err := workspace.runTf(ctx, "plan", "-no-color")
	if err != nil {
		return err
	}
	logger.Info(output.StdOut)

	output, err = workspace.runTf(ctx, "apply", "-auto-approve", "-no-color")
	if err != nil {
		return err
	}
	logger.Info(output.StdOut)

	return nil
}

func (workspace *TerraformWorkspace) MigrateTo10(ctx context.Context) error {
	logger := utils.NewLogger("terraform-migrate-to-1.0").WithData(correlation.ID(ctx))

	err := workspace.initializeFs(ctx)
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	// need to check if flat file system or modules
	if _, err = os.Stat(workspace.dir + "/instance.tf.json"); err == nil {
		err = workspace.AddSensitiveToOutputs()
		if err != nil {
			return err
		}
	}

	output, err := workspace.runTf(ctx, "fmt", "-recursive")
	if err != nil {
		return err
	}
	logger.Info(output.StdOut)

	output, err = workspace.runTf(ctx, "plan", "-no-color")
	if err != nil {
		return err
	}
	logger.Info(output.StdOut)

	output, err = workspace.runTf(ctx, "apply", "-auto-approve", "-no-color")
	if err != nil {
		return err
	}
	logger.Info(output.StdOut)

	return nil
}

func (workspace *TerraformWorkspace) MigrateTo11(ctx context.Context) error {
	logger := utils.NewLogger("terraform-migrate-to-1.1").WithData(correlation.ID(ctx))

	err := workspace.initializeFs(ctx)
	defer workspace.teardownFs()
	if err != nil {
		return err
	}

	output, err := workspace.runTf(ctx, "plan", "-no-color")
	if err != nil {
		return err
	}
	logger.Info(output.StdOut)

	output, err = workspace.runTf(ctx, "apply", "-auto-approve", "-no-color")
	if err != nil {
		return err
	}
	logger.Info(output.StdOut)

	return nil
}

func (workspace *TerraformWorkspace) tfStatePath() string {
	return path.Join(workspace.dir, "terraform.tfstate")
}

func (workspace *TerraformWorkspace) runTf(ctx context.Context, subCommand string, args ...string) (ExecutionOutput, error) {
	sub := []string{subCommand}
	sub = append(sub, args...)

	c := exec.Command("terraform", sub...)
	c.Env = os.Environ()
	c.Dir = workspace.dir

	executor := DefaultExecutor
	if workspace.Executor != nil {
		executor = workspace.Executor
	}

	return executor(ctx, c)
}

type tfState struct {
	Version string `json:"terraform_version"`
}

func (workspace *TerraformWorkspace) StateVersion() (*version.Version, error) {
	tf := tfState{}
	err := json.Unmarshal(workspace.State, &tf)
	if err != nil {
		return nil, err
	}
	return version.NewVersion(tf.Version)

}

// CustomEnvironmentExecutor sets custom environment variables on the Terraform
// execution.
func CustomEnvironmentExecutor(environment map[string]string, wrapped TerraformExecutor) TerraformExecutor {
	return func(ctx context.Context, c *exec.Cmd) (ExecutionOutput, error) {
		for k, v := range environment {
			c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
		}

		return wrapped(ctx, c)
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
func CustomTerraformExecutor(tfBinaryPath, tfPluginDir string, tfVersion *version.Version, wrapped TerraformExecutor) TerraformExecutor {
	return func(ctx context.Context, c *exec.Cmd) (ExecutionOutput, error) {
		subCommand := c.Args[1]
		subCommandArgs := c.Args[2:]

		if subCommand == "init" {
			if tfVersion.LessThan(version.Must(version.NewVersion("0.13.0"))) {
				subCommandArgs = append([]string{"-get-plugins=false"}, subCommandArgs...)
			}
			subCommandArgs = append([]string{fmt.Sprintf("-plugin-dir=%s", tfPluginDir)}, subCommandArgs...)
		}

		allArgs := append([]string{subCommand}, subCommandArgs...)
		newCmd := exec.Command(tfBinaryPath, allArgs...)
		newCmd.Dir = c.Dir
		newCmd.Env = append(c.Env, updatePath(c.Env, tfPluginDir))
		return wrapped(ctx, newCmd)
	}
}

// DefaultExecutor is the default executor that shells out to Terraform
// and logs results to stdout.
func DefaultExecutor(ctx context.Context, c *exec.Cmd) (ExecutionOutput, error) {
	logger := utils.NewLogger("terraform@" + c.Dir).WithData(correlation.ID(ctx))

	logger.Info("starting process", lager.Data{
		"path": c.Path,
		"args": c.Args,
		"dir":  c.Dir,
	})

	stderr, err := c.StderrPipe()
	if err != nil {
		return ExecutionOutput{}, fmt.Errorf("failed to get stderr pipe for terraform execution: %v", err)
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		return ExecutionOutput{}, fmt.Errorf("failed to get stdout pipe for terraform execution: %v", err)
	}

	if err := c.Start(); err != nil {
		return ExecutionOutput{}, fmt.Errorf("failed to execute terraform: %v", err)
	}

	output, _ := io.ReadAll(stdout)
	errors, _ := io.ReadAll(stderr)

	err = c.Wait()

	if err != nil ||
		len(errors) > 0 {
		logger.Error("terraform execution failed", err, lager.Data{
			"errors": string(errors),
		})
	}

	logger.Info("finished process")
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
