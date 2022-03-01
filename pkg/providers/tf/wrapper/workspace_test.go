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
	"context"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"

	"github.com/hashicorp/go-version"
)

func TestTerraformWorkspace_Invariants(t *testing.T) {

	// This function tests the following two invariants of the workspace:
	// - The function updates the tfstate once finished.
	// - The function creates and destroys its own dir.

	cases := map[string]struct {
		Exec func(ws *TerraformWorkspace, executor TerraformExecutor)
	}{
		"validate": {Exec: func(ws *TerraformWorkspace, executor TerraformExecutor) {
			ws.Validate(context.TODO(), executor)
		}},
		"apply": {Exec: func(ws *TerraformWorkspace, executor TerraformExecutor) {
			ws.Apply(context.TODO(), executor)
		}},
		"destroy": {Exec: func(ws *TerraformWorkspace, executor TerraformExecutor) {
			ws.Destroy(context.TODO(), executor)
		}},
		"import": {Exec: func(ws *TerraformWorkspace, executor TerraformExecutor) {
			ws.Import(context.TODO(), executor, map[string]string{})
		}},
		"show": {Exec: func(ws *TerraformWorkspace, executor TerraformExecutor) {
			ws.Show(context.TODO(), executor)
		}},
		"plan": {Exec: func(ws *TerraformWorkspace, executor TerraformExecutor) {
			ws.Plan(context.TODO(), executor)
		}},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			// construct workspace
			const definitionTfContents = "variable azure_tenant_id { type = string }"
			ws, err := NewWorkspace(map[string]interface{}{}, definitionTfContents, map[string]string{}, []ParameterMapping{}, []string{}, []ParameterMapping{})
			if err != nil {
				t.Fatal(err)
			}

			// substitute the executor so we can validate the state at the time of
			// "running" tf
			executorRan := false
			cmdDir := ""
			executor := func(ctx context.Context, cmd *exec.Cmd) (ExecutionOutput, error) {
				executorRan = true
				cmdDir = cmd.Dir

				// validate that the directory exists
				_, err := os.Stat(cmd.Dir)
				if err != nil {
					t.Fatalf("couldn't stat the cmd execution dir %v", err)
				}

				variables, err := os.ReadFile(path.Join(cmd.Dir, "brokertemplate", "definition.tf"))
				if err != nil {
					t.Fatalf("couldn't read the tf file %v", err)
				}
				if string(variables) != definitionTfContents {
					t.Fatalf("Contents of %s should be %s, but got %s", path.Join(cmd.Dir, "brokertemplate", "defintion.tf"), definitionTfContents, string(variables))
				}

				// write dummy state file
				if err := os.WriteFile(path.Join(cmdDir, "terraform.tfstate"), []byte(tn), 0755); err != nil {
					t.Fatal(err)
				}

				return ExecutionOutput{}, nil
			}

			// run function
			tc.Exec(ws, executor)

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
		Exec func(ws *TerraformWorkspace, executor TerraformExecutor)
	}{
		"validate": {Exec: func(ws *TerraformWorkspace, executor TerraformExecutor) {
			ws.Validate(context.TODO(), executor)
		}},
		"apply": {Exec: func(ws *TerraformWorkspace, executor TerraformExecutor) {
			ws.Apply(context.TODO(), executor)
		}},
		"destroy": {Exec: func(ws *TerraformWorkspace, executor TerraformExecutor) {
			ws.Destroy(context.TODO(), executor)
		}},
		"import": {Exec: func(ws *TerraformWorkspace, executor TerraformExecutor) {
			ws.Import(context.TODO(), executor, map[string]string{})
		}},
		"show": {Exec: func(ws *TerraformWorkspace, executor TerraformExecutor) {
			ws.Show(context.TODO(), executor)
		}},
		"plan": {Exec: func(ws *TerraformWorkspace, executor TerraformExecutor) {
			ws.Plan(context.TODO(), executor)
		}},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			// construct workspace
			const variablesTfContents = "variable azure_tenant_id { type = string }"
			ws, err := NewWorkspace(map[string]interface{}{}, ``, map[string]string{"variables": variablesTfContents}, []ParameterMapping{}, []string{}, []ParameterMapping{})
			if err != nil {
				t.Fatal(err)
			}

			// substitute the executor so we can validate the state at the time of
			// "running" tf
			executorRan := false
			cmdDir := ""
			executor := func(ctx context.Context, cmd *exec.Cmd) (ExecutionOutput, error) {
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

				return ExecutionOutput{}, nil
			}

			// run function
			tc.Exec(ws, executor)

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

func TestCustomTerraformExecutor012(t *testing.T) {
	customBinary := "/path/to/terraform"
	customPlugins := "/path/to/terraform-plugins"
	pluginsFlag := "-plugin-dir=" + customPlugins

	cases := map[string]struct {
		Input    *exec.Cmd
		Expected *exec.Cmd
	}{
		"destroy": {
			Input:    exec.Command("terraform", "destroy", "-auto-approve", "-no-color"),
			Expected: exec.Command(customBinary, "destroy", "-auto-approve", "-no-color"),
		},
		"apply": {
			Input:    exec.Command("terraform", "apply", "-auto-approve", "-no-color"),
			Expected: exec.Command(customBinary, "apply", "-auto-approve", "-no-color"),
		},
		"validate": {
			Input:    exec.Command("terraform", "validate", "-no-color"),
			Expected: exec.Command(customBinary, "validate", "-no-color"),
		},
		"init": {
			Input:    exec.Command("terraform", "init", "-no-color"),
			Expected: exec.Command(customBinary, "init", pluginsFlag, "-get-plugins=false", "-no-color"),
		},
		"import": {
			Input:    exec.Command("terraform", "import", "-no-color", "tf.resource", "iaas-resource"),
			Expected: exec.Command(customBinary, "import", "-no-color", "tf.resource", "iaas-resource"),
		},
		"show": {
			Input:    exec.Command("terraform", "show", "-no-color"),
			Expected: exec.Command(customBinary, "show", "-no-color"),
		},
		"plan": {
			Input:    exec.Command("terraform", "plan", "-no-color"),
			Expected: exec.Command(customBinary, "plan", "-no-color"),
		},
	}

	for tn, tc := range cases {
		tc.Input.Env = []string{"PATH=/foo", "ENV1=bar"}
		tc.Expected.Env = []string{"PATH=/foo", "ENV1=bar", "PATH=/path/to/terraform-plugins:/foo"}
		t.Run(tn, func(t *testing.T) {
			actual := exec.Command("!actual-never-got-called!")

			tfVersion, _ := version.NewVersion("0.12.0")
			executor := CustomTerraformExecutor(customBinary, customPlugins, tfVersion, func(ctx context.Context, c *exec.Cmd) (ExecutionOutput, error) {
				actual = c
				return ExecutionOutput{}, nil
			})

			executor(context.TODO(), tc.Input)

			if actual.Path != tc.Expected.Path {
				t.Errorf("path wasn't updated, expected: %q, actual: %q", tc.Expected.Path, actual.Path)
			}

			if !reflect.DeepEqual(actual.Env, tc.Expected.Env) {
				t.Errorf("env wasn't updated, expected: %q, actual: %q", tc.Expected.Env, actual.Env)
			}

			if !reflect.DeepEqual(actual.Args, tc.Expected.Args) {
				t.Errorf("args weren't updated correctly, expected: %#v, actual: %#v", tc.Expected.Args, actual.Args)
			}
		})
	}
}

func TestCustomTerraformExecutor013(t *testing.T) {
	customBinary := "/path/to/terraform"
	customPlugins := "/path/to/terraform-plugins"
	pluginsFlag := "-plugin-dir=" + customPlugins

	cases := map[string]struct {
		Input    *exec.Cmd
		Expected *exec.Cmd
	}{
		"init": {
			Input:    exec.Command("terraform", "init", "-no-color"),
			Expected: exec.Command(customBinary, "init", pluginsFlag, "-no-color"),
		},
	}

	for tn, tc := range cases {
		tc.Input.Env = []string{"PATH=/foo", "ENV1=bar"}
		tc.Expected.Env = []string{"PATH=/foo", "ENV1=bar", "PATH=/path/to/terraform-plugins:/foo"}
		t.Run(tn, func(t *testing.T) {
			actual := exec.Command("!actual-never-got-called!")

			tfVersion, _ := version.NewVersion("0.13.0")
			executor := CustomTerraformExecutor(customBinary, customPlugins, tfVersion, func(ctx context.Context, c *exec.Cmd) (ExecutionOutput, error) {
				actual = c
				return ExecutionOutput{}, nil
			})

			executor(context.TODO(), tc.Input)

			if actual.Path != tc.Expected.Path {
				t.Errorf("path wasn't updated, expected: %q, actual: %q", tc.Expected.Path, actual.Path)
			}

			if !reflect.DeepEqual(actual.Env, tc.Expected.Env) {
				t.Errorf("env wasn't updated, expected: %q, actual: %q", tc.Expected.Env, actual.Env)
			}

			if !reflect.DeepEqual(actual.Args, tc.Expected.Args) {
				t.Errorf("args weren't updated correctly, expected: %#v, actual: %#v", tc.Expected.Args, actual.Args)
			}
		})
	}
}

func TestCustomEnvironmentExecutor(t *testing.T) {
	c := exec.Command("/path/to/terraform", "apply")
	c.Env = []string{"ORIGINAL=value"}

	actual := exec.Command("!actual-never-got-called!")
	executor := CustomEnvironmentExecutor(map[string]string{"FOO": "bar"}, func(ctx context.Context, c *exec.Cmd) (ExecutionOutput, error) {
		actual = c
		return ExecutionOutput{}, nil
	})

	executor(context.TODO(), c)
	expected := []string{"ORIGINAL=value", "FOO=bar"}

	if !reflect.DeepEqual(expected, actual.Env) {
		t.Fatalf("Expected %v actual %v", expected, actual)
	}
}
