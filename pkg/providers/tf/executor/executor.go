// Package executor executes Terraform
package executor

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-version"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . TerraformExecutor

// TerraformExecutor is the function that shells out to Terraform.
// It can intercept, modify or retry the given command.
type TerraformExecutor interface {
	Execute(context.Context, *exec.Cmd) (ExecutionOutput, error)
}

// ExecutionOutput captures output from tf cli execution
type ExecutionOutput struct {
	StdOut string
	StdErr string
}

// DefaultExecutor is the default executor that shells out to Terraform
// and logs results to stdout.
func DefaultExecutor() TerraformExecutor {
	return defaultExecutor{}
}

type defaultExecutor struct{}

func (defaultExecutor) Execute(ctx context.Context, c *exec.Cmd) (ExecutionOutput, error) {
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
		return ExecutionOutput{}, fmt.Errorf("%s %w", flatten(errors), err)
	}

	return ExecutionOutput{
		StdErr: string(errors),
		StdOut: string(output),
	}, nil
}

func flatten(input []byte) string {
	var lines []string
	for _, l := range strings.Split(string(input), "\n") {
		if line := strings.TrimSpace(l); len(line) > 0 {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, " ")
}

// CustomTerraformExecutor executes a custom Terraform binary that uses plugins
// from a given plugin directory rather than the Terraform that's on the PATH
// which will download provider binaries from the web.
func CustomTerraformExecutor(tfBinaryPath, tfPluginDir string, tfVersion *version.Version, wrapped TerraformExecutor) TerraformExecutor {
	return customExecutor{tfBinaryPath: tfBinaryPath, tfPluginDir: tfPluginDir, tfVersion: tfVersion, wrapped: wrapped}
}

type customExecutor struct {
	tfBinaryPath string
	tfPluginDir  string
	tfVersion    *version.Version
	wrapped      TerraformExecutor
}

func (e customExecutor) Execute(ctx context.Context, c *exec.Cmd) (ExecutionOutput, error) {
	newCmd := exec.Command(e.tfBinaryPath, c.Args[1:]...)
	newCmd.Dir = c.Dir
	newCmd.Env = append(c.Env, updatePath(c.Env, e.tfPluginDir))
	return e.wrapped.Execute(ctx, newCmd)
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

// CustomEnvironmentExecutor sets custom environment variables on the Terraform
// execution.
func CustomEnvironmentExecutor(environment map[string]string, wrapped TerraformExecutor) TerraformExecutor {
	return customEnvironmentExecutor{environment: environment, wrapped: wrapped}
}

type customEnvironmentExecutor struct {
	wrapped     TerraformExecutor
	environment map[string]string
}

func (executor customEnvironmentExecutor) Execute(ctx context.Context, c *exec.Cmd) (ExecutionOutput, error) {
	for k, v := range executor.environment {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
	}

	return executor.wrapped.Execute(ctx, c)
}
