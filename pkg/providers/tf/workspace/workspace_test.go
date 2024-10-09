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

package workspace

import (
	"context"
	"errors"
	"maps"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/executor"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/command"

	"github.com/hashicorp/go-version"
)

func TestTerraformWorkspace_Invariants(t *testing.T) {

	// This function tests the following two invariants of the workspace:
	// - The function updates the tfstate once finished.
	// - The function creates and destroys its own dir.

	cases := map[string]struct {
		Exec func(ws *TerraformWorkspace, executor executor.TerraformExecutor)
	}{
		"execute": {Exec: func(ws *TerraformWorkspace, executor executor.TerraformExecutor) {
			_, _ = ws.Execute(context.TODO(), executor, command.NewApply())
		}},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			// construct workspace
			const definitionTfContents = "variable azure_tenant_id { type = string }"
			ws, err := NewWorkspace(map[string]any{}, definitionTfContents, map[string]string{}, []ParameterMapping{}, []string{}, []ParameterMapping{})
			if err != nil {
				t.Fatal(err)
			}

			// substitute the executor, so we can validate the state at the time of
			// "running" tf
			executorRan := false
			cmdDir := ""
			tExecutor := newTestExecutor(func(ctx context.Context, cmd *exec.Cmd) (executor.ExecutionOutput, error) {
				executorRan = true
				cmdDir = cmd.Dir

				// validate that the directory exists
				_, err := os.Stat(cmd.Dir)
				if err != nil {
					t.Fatalf("couldn't stat the cmd execution dir %v", err)
				}

				tfDefinitionFilePath := path.Join(cmd.Dir, "brokertemplate", "definition.tf")
				variables, err := os.ReadFile(tfDefinitionFilePath)
				if err != nil {
					t.Fatalf("couldn't read the tf file %v", err)
				}
				if string(variables) != definitionTfContents {
					t.Fatalf("Contents of %s should be %s, but got %s", tfDefinitionFilePath, definitionTfContents, string(variables))
				}

				// write dummy state file
				if err := os.WriteFile(path.Join(cmdDir, "terraform.tfstate"), []byte(tn), 0755); err != nil {
					t.Fatal(err)
				}

				return executor.ExecutionOutput{}, nil
			})

			// run function
			tc.Exec(ws, tExecutor)

			// check validator got ran
			if !executorRan {
				t.Fatal("executor did not get run as part of the function")
			}

			// check workspace destroyed
			if _, err := os.Stat(cmdDir); !os.IsNotExist(err) {
				t.Fatalf("command directory didn't %q get torn down %v", cmdDir, err)
			}

			// check tfstate updated
			if !reflect.DeepEqual(ws.State, []byte(tn)) {
				t.Fatalf("Expected state %v got %v", []byte(tn), ws.State)
			}
		})
	}
}

func TestTerraformWorkspace_InvariantsFlat(t *testing.T) {

	// This function tests the following two invariants of the workspace:
	// - The function updates the tfstate once finished.
	// - The function creates and destroys its own dir.

	cases := map[string]struct {
		Exec func(ws *TerraformWorkspace, executor executor.TerraformExecutor)
	}{
		"execute": {Exec: func(ws *TerraformWorkspace, executor executor.TerraformExecutor) {
			_, _ = ws.Execute(context.TODO(), executor, command.NewApply())
		}},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			// construct workspace
			const variablesTfContents = "variable azure_tenant_id { type = string }"
			ws, err := NewWorkspace(map[string]any{}, ``, map[string]string{"variables": variablesTfContents}, []ParameterMapping{}, []string{}, []ParameterMapping{})
			if err != nil {
				t.Fatal(err)
			}

			// substitute the executor, so we can validate the state at the time of
			// "running" tf
			executorRan := false
			cmdDir := ""
			tExecutor := newTestExecutor(func(ctx context.Context, cmd *exec.Cmd) (executor.ExecutionOutput, error) {
				executorRan = true
				cmdDir = cmd.Dir

				// validate that the directory exists
				_, err := os.Stat(cmd.Dir)
				if err != nil {
					t.Fatalf("couldn't stat the cmd execution dir %v", err)
				}

				variables, err := os.ReadFile(path.Join(cmd.Dir, "variables.tf"))
				if err != nil {
					t.Fatalf("couldn't read the tf file %v", err)
				}
				if string(variables) != variablesTfContents {
					t.Fatalf("Contents of %s should be %s, but got %s", path.Join(cmd.Dir, "brokertemplate", "variables.tf"), variablesTfContents, string(variables))
				}

				// write dummy state file
				if err := os.WriteFile(path.Join(cmdDir, "terraform.tfstate"), []byte(tn), 0755); err != nil {
					t.Fatal(err)
				}

				return executor.ExecutionOutput{}, nil
			})

			// run function
			tc.Exec(ws, tExecutor)

			// check validator got ran
			if !executorRan {
				t.Fatal("executor did not get run as part of the function")
			}

			// check workspace destroyed
			if _, err := os.Stat(cmdDir); !os.IsNotExist(err) {
				t.Fatalf("command directory didn't %q get torn down %v", cmdDir, err)
			}

			// check tfstate updated
			if !reflect.DeepEqual(ws.State, []byte(tn)) {
				t.Fatalf("Expected state %v got %v", []byte(tn), ws.State)
			}
		})
	}
}

func TestTerrafromWorkspace_StateTFVersion(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		e := version.Must(version.NewVersion("1.2.3"))
		w := &TerraformWorkspace{State: []byte(`{"terraform_version":"1.2.3"}`)}
		v, err := w.StateTFVersion()
		switch {
		case err != nil:
			t.Fatalf("unexpected error: %s", err)
		case !v.Equal(e):
			t.Fatalf("wrong version, expected %q, got %q", e, v)
		}
	})

	t.Run("empty", func(t *testing.T) {
		e := CannotReadVersionError{message: "workspace state not generated"}
		w := &TerraformWorkspace{State: nil}
		_, err := w.StateTFVersion()
		if !errors.Is(err, e) {
			t.Fatalf("wrong error type, expected: %T; got: %T", e, err)
		}
	})

	t.Run("json", func(t *testing.T) {
		e := CannotReadVersionError{message: "invalid workspace state unexpected end of JSON input"}
		w := &TerraformWorkspace{State: []byte(`{"foo`)}
		_, err := w.StateTFVersion()
		if !errors.Is(err, e) {
			t.Fatalf("wrong error, expected: %q %T; got: %q %T", e, e, err, err)
		}
	})
}

func TestTerrafromWorkspace_Execute(t *testing.T) {
	t.Parallel()

	t.Run("custom command env", func(t *testing.T) {
		const definitionTfContents = "variable azure_tenant_id { type = string }"
		ws, err := NewWorkspace(map[string]any{}, definitionTfContents, map[string]string{}, []ParameterMapping{}, []string{}, []ParameterMapping{})
		if err != nil {
			t.Fatal(err)
		}

		tExecutor := newTestExecutor(func(ctx context.Context, cmd *exec.Cmd) (executor.ExecutionOutput, error) {
			if cmd.Env[len(cmd.Env)-1] != "OPENTOFU_STATEFILE_PROVIDER_ADDRESS_TRANSLATION=0" {
				t.Fatalf("Custom command environment variable not set. Expected `OPENTOFU_STATEFILE_PROVIDER_ADDRESS_TRANSLATION=0` got %s", cmd.Env[len(cmd.Env)-1])
			}
			return executor.ExecutionOutput{}, nil
		})

		ws.Execute(context.TODO(), tExecutor, command.NewShow())

	})

	t.Run("empty command env", func(t *testing.T) {
		const definitionTfContents = "variable azure_tenant_id { type = string }"
		ws, err := NewWorkspace(map[string]any{}, definitionTfContents, map[string]string{}, []ParameterMapping{}, []string{}, []ParameterMapping{})
		if err != nil {
			t.Fatal(err)
		}

		tExecutor := newTestExecutor(func(ctx context.Context, cmd *exec.Cmd) (executor.ExecutionOutput, error) {
			if !reflect.DeepEqual(cmd.Env, os.Environ()) {
				t.Fatalf("Unexpected env variable set. Expected %s got %s", os.Environ(), cmd.Env)
			}
			return executor.ExecutionOutput{}, nil
		})

		ws.Execute(context.TODO(), tExecutor, command.NewApply())

	})
}

func TestCustomEnvironmentExecutor(t *testing.T) {
	c := exec.Command("/path/to/terraform", "apply")
	c.Env = []string{"ORIGINAL=value"}

	actual := exec.Command("!actual-never-got-called!")
	customEnvExecutor := executor.CustomEnvironmentExecutor(map[string]string{"FOO": "bar"}, newTestExecutor(func(ctx context.Context, c *exec.Cmd) (executor.ExecutionOutput, error) {
		actual = c
		return executor.ExecutionOutput{}, nil
	}))

	_, _ = customEnvExecutor.Execute(context.TODO(), c)
	expected := []string{"ORIGINAL=value", "FOO=bar"}

	if !reflect.DeepEqual(expected, actual.Env) {
		t.Fatalf("Expected %v actual %v", expected, actual)
	}
}
func newTestExecutor(function func(ctx context.Context, cmd *exec.Cmd) (executor.ExecutionOutput, error)) testExecutor {
	return testExecutor{function: function}
}

type testExecutor struct {
	function func(ctx context.Context, cmd *exec.Cmd) (executor.ExecutionOutput, error)
}

func (exec testExecutor) Execute(ctx context.Context, c *exec.Cmd) (executor.ExecutionOutput, error) {
	return exec.function(ctx, c)
}

func TestNewWorkspace(t *testing.T) {
	// NewWorkspace creates deep copies of data which is read from static service definitions.

	t.Parallel()

	t.Run("deep copy terraformTemplates", func(t *testing.T) {

		in := map[string]string{
			"main.tf": "variable domain {type = string}\nvariable username {type = string}\noutput email {value = \"${var.username}@${var.domain}\"}",
		}

		inCopy := make(map[string]string)
		maps.Copy(inCopy, in)

		ws, err := NewWorkspace(nil, "", in, nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		ws.Modules[0].Definitions["main.tf"] = ""

		if !reflect.DeepEqual(in, inCopy) {
			t.Error("Expected NewWorkspace to create deep copy of terraformTemplates")
		}
	})

	t.Run("deep copy importParameterMappings", func(t *testing.T) {
		in := []ParameterMapping{
			{
				TfVariable:    "aws_s3_bucket.this.bucket",
				ParameterName: "bucket_name",
			},
		}

		inCopy := make([]ParameterMapping, len(in))
		copy(inCopy, in)

		ws, err := NewWorkspace(nil, "", nil, in, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		ws.Transformer.ParameterMappings[0] = ParameterMapping{}

		if !reflect.DeepEqual(in, inCopy) {
			t.Error("Expected NewWorkspace to create deep copy of importParameterMappings")
		}
	})

	t.Run("deep copy parametersToRemove", func(t *testing.T) {
		in := []string{
			"aws_s3_bucket.this.id",
		}

		inCopy := make([]string, len(in))
		copy(inCopy, in)

		ws, err := NewWorkspace(nil, "", nil, nil, in, nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		ws.Transformer.ParametersToRemove[0] = ""

		if !reflect.DeepEqual(in, inCopy) {
			t.Error("Expected NewWorkspace to create deep copy of parametersToRemove")
		}
	})

	t.Run("deep copy parametersToAdd", func(t *testing.T) {
		in := []ParameterMapping{
			{
				TfVariable:    "aws_s3_bucket.this.bucket_domain_name",
				ParameterName: "domain_name",
			},
		}

		inCopy := make([]ParameterMapping, len(in))
		copy(inCopy, in)

		ws, err := NewWorkspace(nil, "", nil, nil, nil, in)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		ws.Transformer.ParametersToAdd[0] = ParameterMapping{}

		if !reflect.DeepEqual(in, inCopy) {
			t.Error("Expected NewWorkspace to create deep copy of ParametersToAdd")
		}
	})
	t.Run("returns empty map for empty templateVars", func(t *testing.T) {
		ws, err := NewWorkspace(nil, "", nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		got := ws.Instances[0].Configuration
		expect := make(map[string]any)
		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expected %v, got %v", expect, got)
		}
	})

	t.Run("returns zero-value for empty terraformTemplate", func(t *testing.T) {
		ws, err := NewWorkspace(nil, "", nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		got := ws.Modules[0].Definition
		var expect string
		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expected %v, got %v", expect, got)
		}
	})

	t.Run("returns zero-value for empty terraformTemplates", func(t *testing.T) {
		ws, err := NewWorkspace(nil, "", nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		got := ws.Modules[0].Definitions
		var expect map[string]string
		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expected %v, got %v", expect, got)
		}
	})

	t.Run("returns zero-value for empty importParameterMappings", func(t *testing.T) {
		ws, err := NewWorkspace(nil, "", nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		got := ws.Transformer.ParameterMappings
		var expect []ParameterMapping
		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expected %v, got %v", expect, got)
		}
	})

	t.Run("returns zero-value for empty importParametersToRemove", func(t *testing.T) {
		ws, err := NewWorkspace(nil, "", nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		got := ws.Transformer.ParametersToRemove
		var expect []string
		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expected %v, got %v", expect, got)
		}
	})

	t.Run("returns zero-value for empty importParametersToAdd", func(t *testing.T) {
		ws, err := NewWorkspace(nil, "", nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		got := ws.Transformer.ParametersToAdd
		var expect []ParameterMapping
		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expected %v, got %v", expect, got)
		}
	})

}
