package wrapper

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-version"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

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

// CustomTerraformExecutor executes a custom Terraform binary that uses plugins
// from a given plugin directory rather than the Terraform that's on the PATH
// which will download provider binaries from the web.
func CustomTerraformExecutor(tfBinaryPath, tfPluginDir string, tfVersion *version.Version, wrapped TerraformExecutor) TerraformExecutor {
	return customerExecutor{tfBinaryPath: tfBinaryPath, tfPluginDir: tfPluginDir, tfVersion: tfVersion, wrapped: wrapped}
}

type customerExecutor struct {
	tfBinaryPath string
	tfPluginDir  string
	tfVersion    *version.Version
	wrapped      TerraformExecutor
}

func (e customerExecutor) TerraformExecutor(ctx context.Context, c *exec.Cmd) (ExecutionOutput, error) {
	subCommand := c.Args[1]
	subCommandArgs := c.Args[2:]

	if subCommand == "init" {
		if e.tfVersion.LessThan(version.Must(version.NewVersion("0.13.0"))) {
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
