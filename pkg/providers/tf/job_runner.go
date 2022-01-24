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

package tf

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-version"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf/wrapper"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/correlation"
)

const (
	InProgress = "in progress"
	Succeeded  = "succeeded"
	Failed     = "failed"
)

// NewTfJobRunner constructs a new JobRunner for the given project.
func NewTfJobRunner(envVars map[string]string, store broker.ServiceProviderStorage) *TfJobRunner {
	return &TfJobRunner{
		EnvVars: envVars,
		store:   store,
	}
}

// TfJobRunner is responsible for executing terraform jobs in the background and
// providing a way to log and access the state of those background tasks.
//
// Jobs are given an ID and a workspace to operate in, and then the TfJobRunner
// is told which Terraform commands to execute on the given job.
// The TfJobRunner executes those commands in the background and keeps track of
// their state in a database table which gets updated once the task is completed.
//
// The TfJobRunner keeps track of the workspace and the Terraform state file so
// subsequent commands will operate on the same structure.
type TfJobRunner struct {
	// EnvVars is a list of environment variables that should be included in executor
	// env (usually Terraform provider credentials)
	EnvVars map[string]string
	// Executor holds a custom executor that will be called when commands are run.
	Executor wrapper.TerraformExecutor
	store    broker.ServiceProviderStorage
}

// StageJob stages a job to be executed. Before the workspace is saved to the
// database, the modules and inputs are validated by Terraform.
func (runner *TfJobRunner) StageJob(jobId string, workspace *wrapper.TerraformWorkspace) error {
	workspace.Executor = runner.Executor

	deployment := storage.TerraformDeployment{ID: jobId}
	exists, err := runner.store.ExistsTerraformDeployment(jobId)
	switch {
	case err != nil:
		return err
	case exists:
		deployment, err = runner.store.GetTerraformDeployment(jobId)
		if err != nil {
			return err
		}
	default:
		if err := runner.store.StoreTerraformDeployment(deployment); err != nil {
			return err
		}
	}

	workspaceString, err := workspace.Serialize()
	if err != nil {
		return err
	}

	deployment.Workspace = []byte(workspaceString)
	deployment.LastOperationType = "validation"
	return runner.operationFinished(nil, workspace, deployment)
}

func (runner *TfJobRunner) markJobStarted(deployment storage.TerraformDeployment, operationType string) error {
	// update the deployment info
	deployment.LastOperationType = operationType
	deployment.LastOperationState = InProgress
	deployment.LastOperationMessage = ""

	if err := runner.store.StoreTerraformDeployment(deployment); err != nil {
		return err
	}

	return nil
}

func (runner *TfJobRunner) hydrateWorkspace(ctx context.Context, deployment storage.TerraformDeployment) (*wrapper.TerraformWorkspace, error) {
	ws, err := wrapper.DeserializeWorkspace(string(deployment.Workspace))
	if err != nil {
		return nil, err
	}

	ws.Executor = wrapper.CustomEnvironmentExecutor(runner.EnvVars, runner.Executor)

	logger := utils.NewLogger("job-runner")
	logger.Debug("wrapping", correlation.ID(ctx), lager.Data{
		"wrapper": ws,
	})

	return ws, nil
}

// ImportResource represents TF resource to IaaS resource ID mapping for import
type ImportResource struct {
	TfResource   string
	IaaSResource string
}

// Import runs `terraform import` and `terraform apply` on the given workspace in the background.
// The status of the job can be found by polling the Status function.
func (runner *TfJobRunner) Import(ctx context.Context, id string, importResources []ImportResource) error {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return err
	}

	workspace, err := runner.hydrateWorkspace(ctx, deployment)
	if err != nil {
		return err
	}

	if err := runner.markJobStarted(deployment, models.ProvisionOperationType); err != nil {
		return err
	}

	go func() {
		logger := utils.NewLogger("Import").WithData(correlation.ID(ctx))
		resources := make(map[string]string)
		for _, resource := range importResources {
			resources[resource.TfResource] = resource.IaaSResource
		}
		if err := workspace.Import(ctx, resources); err != nil {
			logger.Error("Import Failed", err)
			runner.operationFinished(err, workspace, deployment)
			return
		}
		mainTf, err := workspace.Show(ctx)
		if err == nil {
			var tf string
			var parameterVals map[string]string
			tf, parameterVals, err = workspace.Transformer.ReplaceParametersInTf(workspace.Transformer.AddParametersInTf(workspace.Transformer.CleanTf(mainTf)))
			if err == nil {
				for pn, pv := range parameterVals {
					workspace.Instances[0].Configuration[pn] = pv
				}
				workspace.Modules[0].Definitions["main"] = tf

				logger.Info("new workspace", lager.Data{
					"workspace": workspace,
					"tf":        tf,
				})

				err = workspace.Plan(ctx)
				if err != nil {
					logger.Error("plan failed", err)
				} else {
					err = workspace.Apply(ctx)
				}
			}
		}
		runner.operationFinished(err, workspace, deployment)
	}()

	return nil
}

// Create runs `terraform apply` on the given workspace in the background.
// The status of the job can be found by polling the Status function.
func (runner *TfJobRunner) Create(ctx context.Context, id string) error {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return fmt.Errorf("error getting TF deployment: %w", err)
	}

	workspace, err := runner.hydrateWorkspace(ctx, deployment)
	if err != nil {
		return fmt.Errorf("error hydrating workspace: %w", err)
	}

	if err := runner.markJobStarted(deployment, models.ProvisionOperationType); err != nil {
		return fmt.Errorf("error marking job started: %w", err)
	}

	go func() {
		err := workspace.Apply(ctx)
		runner.operationFinished(err, workspace, deployment)
	}()

	return nil
}

func (runner *TfJobRunner) Update(ctx context.Context, id string, templateVars map[string]interface{}) error {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return err
	}

	workspace, err := runner.hydrateWorkspace(ctx, deployment)
	if err != nil {
		return err
	}

	// we may be doing this twice in the case of dynamic HCL, that is fine.
	inputList, err := workspace.Modules[0].Inputs()
	if err != nil {
		return err
	}

	limitedConfig := make(map[string]interface{})
	for _, name := range inputList {
		limitedConfig[name] = templateVars[name]
	}

	workspace.Instances[0].Configuration = limitedConfig

	if err := runner.markJobStarted(deployment, models.UpdateOperationType); err != nil {
		return err
	}

	go func() {
		err := workspace.Apply(ctx)
		runner.operationFinished(err, workspace, deployment)
	}()

	return nil
}

// Destroy runs `terraform destroy` on the given workspace in the background.
// The status of the job can be found by polling the Status function.
func (runner *TfJobRunner) Destroy(ctx context.Context, id string, templateVars map[string]interface{}) error {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return err
	}

	workspace, err := runner.hydrateWorkspace(ctx, deployment)
	if err != nil {
		return err
	}

	inputList, err := workspace.Modules[0].Inputs()
	if err != nil {
		return err
	}

	limitedConfig := make(map[string]interface{})
	for _, name := range inputList {
		limitedConfig[name] = templateVars[name]
	}

	workspace.Instances[0].Configuration = limitedConfig

	if err := runner.markJobStarted(deployment, models.DeprovisionOperationType); err != nil {
		return err
	}

	go func() {
		err := workspace.Destroy(ctx)
		runner.operationFinished(err, workspace, deployment)
	}()

	return nil
}

// operationFinished closes out the state of the background job so clients that
// are polling can get the results.
func (runner *TfJobRunner) operationFinished(err error, workspace *wrapper.TerraformWorkspace, deployment storage.TerraformDeployment) error {
	// we shouldn't update the status on update when updating the HCL, as the status comes either from the provision call or a previous update
	if err == nil {
		lastOperationMessage := ""
		// maybe do if deployment.LastOperationType != "validation" so we don't do the status update on staging a job.
		// previously we would only stage a job on provision so state would be empty and the outputs would be null.
		outputs, err := workspace.Outputs(workspace.Instances[0].InstanceName)
		if err == nil {
			if status, ok := outputs["status"]; ok {
				lastOperationMessage = fmt.Sprintf("%v", status)
			}
		}
		deployment.LastOperationState = Succeeded
		deployment.LastOperationMessage = lastOperationMessage
	} else {
		deployment.LastOperationState = Failed
		deployment.LastOperationMessage = err.Error()
	}

	workspaceString, err := workspace.Serialize()
	if err != nil {
		deployment.LastOperationState = Failed
		deployment.LastOperationMessage = fmt.Sprintf("couldn't serialize workspace, contact your operator for cleanup: %s", err.Error())
	}

	deployment.Workspace = []byte(workspaceString)

	return runner.store.StoreTerraformDeployment(deployment)
}

// Status gets the status of the most recent job on the workspace.
// If isDone is true, then the status of the operation will not change again.
// if isDone is false, then the operation is ongoing.
func (runner *TfJobRunner) Status(ctx context.Context, id string) (bool, string, error) {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return true, "", err
	}

	switch deployment.LastOperationState {
	case Succeeded:
		return true, deployment.LastOperationMessage, nil
	case Failed:
		return true, deployment.LastOperationMessage, errors.New(deployment.LastOperationMessage)
	default:
		return false, deployment.LastOperationMessage, nil
	}
}

// Outputs gets the output variables for the given module instance in the workspace.
func (runner *TfJobRunner) Outputs(ctx context.Context, id, instanceName string) (map[string]interface{}, error) {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return nil, fmt.Errorf("error getting TF deployment: %w", err)
	}

	ws, err := wrapper.DeserializeWorkspace(string(deployment.Workspace))
	if err != nil {
		return nil, fmt.Errorf("error deserializing workspace: %w", err)
	}

	return ws.Outputs(instanceName)
}

// Wait waits for an operation to complete, polling its status once per second.
func (runner *TfJobRunner) Wait(ctx context.Context, id string) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		case <-time.After(1 * time.Second):
			isDone, _, err := runner.Status(ctx, id)
			if isDone {
				return err
			}
		}
	}
}

func (runner *TfJobRunner) MigrateTo013(ctx context.Context, id string) error {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return err
	}

	workspace, err := runner.hydrateWorkspace(ctx, deployment)
	if err != nil {
		return err
	}
	stateVersion, err := workspace.StateVersion()
	if err != nil {
		return err
	}
	if stateVersion.GreaterThan(version.Must(version.NewVersion("0.13.7"))) {
		return nil
	}

	err = workspace.MigrateTo013(ctx)
	if err != nil {
		return err
	}

	workspaceString, err := workspace.Serialize()
	if err != nil {
		return err
	}
	deployment.Workspace = []byte(workspaceString)

	return runner.store.StoreTerraformDeployment(deployment)
}

func (runner *TfJobRunner) MigrateTo014(ctx context.Context, id string) error {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return err
	}

	workspace, err := runner.hydrateWorkspace(ctx, deployment)
	if err != nil {
		return err
	}
	stateVersion, err := workspace.StateVersion()
	if err != nil {
		return err
	}
	if stateVersion.GreaterThan(version.Must(version.NewVersion("0.14.9"))) {
		return nil
	}

	err = workspace.MigrateTo014(ctx)
	if err != nil {
		return err
	}

	workspaceString, err := workspace.Serialize()
	if err != nil {
		return err
	}
	deployment.Workspace = []byte(workspaceString)

	return runner.store.StoreTerraformDeployment(deployment)
}

func (runner *TfJobRunner) MigrateTo10(ctx context.Context, id string) error {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return err
	}

	workspace, err := runner.hydrateWorkspace(ctx, deployment)
	if err != nil {
		return err
	}

	stateVersion, err := workspace.StateVersion()
	if err != nil {
		return err
	}
	if stateVersion.GreaterThan(version.Must(version.NewVersion("1.0.9"))) {
		return nil
	}

	err = workspace.MigrateTo10(ctx)
	if err != nil {
		return err
	}

	workspaceString, err := workspace.Serialize()
	if err != nil {
		return err
	}
	deployment.Workspace = []byte(workspaceString)

	return runner.store.StoreTerraformDeployment(deployment)
}

func (runner *TfJobRunner) MigrateTo11(ctx context.Context, id string) error {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return err
	}

	workspace, err := runner.hydrateWorkspace(ctx, deployment)
	if err != nil {
		return err
	}

	stateVersion, err := workspace.StateVersion()
	if err != nil {
		return err
	}

	if stateVersion.GreaterThan(version.Must(version.NewVersion("1.1.0"))) {
		return nil
	}

	err = workspace.MigrateTo11(ctx)
	if err != nil {
		return err
	}

	workspaceString, err := workspace.Serialize()
	if err != nil {
		return err
	}
	deployment.Workspace = []byte(workspaceString)

	return runner.store.StoreTerraformDeployment(deployment)
}
