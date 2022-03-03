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

	"github.com/spf13/viper"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

const (
	InProgress = "in progress"
	Succeeded  = "succeeded"
	Failed     = "failed"

	TfUpgradeEnabled = "brokerpak.terraform.upgrades.enabled"
)

func init() {
	viper.BindEnv(TfUpgradeEnabled, "TERRAFORM_UPGRADES_ENABLED")
	viper.SetDefault(TfUpgradeEnabled, false)
}

// NewTfJobRunner constructs a new JobRunner for the given project.
func NewTfJobRunner(store broker.ServiceProviderStorage, executorFactory wrapper.ExecutorBuilder, tfBinContext wrapper.TFBinariesContext, workspaceFactory WorkspaceBuilder) *TfJobRunner {
	return &TfJobRunner{
		store:            store,
		tfBinContext:     tfBinContext,
		WorkspaceBuilder: workspaceFactory,
		ExecutorBuilder:  executorFactory,
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
	// executor holds a custom executor that will be called when commands are run.
	store        broker.ServiceProviderStorage
	tfBinContext wrapper.TFBinariesContext
	WorkspaceBuilder
	wrapper.ExecutorBuilder
}

// StageJob stages a job to be executed. Before the workspace is saved to the
// database, the modules and inputs are validated by Terraform.
func (runner *TfJobRunner) StageJob(jobId string, workspace *wrapper.TerraformWorkspace) error {
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
	}

	workspaceString, err := workspace.Serialize()
	if err != nil {
		return err
	}

	deployment.Workspace = []byte(workspaceString)
	deployment.LastOperationType = "validation"

	return runner.store.StoreTerraformDeployment(deployment)
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

	workspace, err := wrapper.DeserializeWorkspace(deployment.Workspace)
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
		if err := workspace.Import(ctx, runner.DefaultExecutor(), resources); err != nil {
			logger.Error("Import Failed", err)
			runner.operationFinished(err, workspace, deployment)
			return
		}
		mainTf, err := workspace.Show(ctx, runner.DefaultExecutor())
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

				err = workspace.Plan(ctx, runner.DefaultExecutor())
				if err != nil {
					logger.Error("plan failed", err)
				} else {
					err = workspace.Apply(ctx, runner.DefaultExecutor())
				}
			}
		}
		runner.operationFinished(err, workspace, deployment)
	}()

	return nil
}

func (runner *TfJobRunner) DefaultExecutor() wrapper.TerraformExecutor {
	return runner.VersionedExecutor(runner.tfBinContext.DefaultTfVersion)
}

// Create runs `terraform apply` on the given workspace in the background.
// The status of the job can be found by polling the Status function.
func (runner *TfJobRunner) Create(ctx context.Context, id string) error {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return fmt.Errorf("error getting TF deployment: %w", err)
	}

	workspace, err := wrapper.DeserializeWorkspace(deployment.Workspace)
	if err != nil {
		return fmt.Errorf("error hydrating workspace: %w", err)
	}

	if err := runner.markJobStarted(deployment, models.ProvisionOperationType); err != nil {
		return fmt.Errorf("error marking job started: %w", err)
	}

	go func() {
		err := workspace.Apply(ctx, runner.DefaultExecutor())
		runner.operationFinished(err, workspace, deployment)
	}()

	return nil
}

func (runner *TfJobRunner) Update(ctx context.Context, id string, templateVars map[string]interface{}) error {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return err
	}

	workspace, err := runner.CreateWorkspace(deployment)
	if err != nil {
		return err
	}

	if err := runner.markJobStarted(deployment, models.UpdateOperationType); err != nil {
		return err
	}

	go func() {
		err = runner.performTerraformUpgrade(ctx, workspace)
		if err != nil {
			runner.operationFinished(err, workspace, deployment)
			return
		}
		err = workspace.UpdateInstanceConfiguration(templateVars)
		if err != nil {
			runner.operationFinished(err, workspace, deployment)
			return
		}

		err = workspace.Apply(ctx, runner.DefaultExecutor())
		runner.operationFinished(err, workspace, deployment)
	}()

	return nil
}

func (runner *TfJobRunner) performTerraformUpgrade(ctx context.Context, workspace Workspace) error {
	currentTfVersion, err := workspace.StateVersion()
	if err != nil {
		return err
	}

	if viper.GetBool(TfUpgradeEnabled) {
		if currentTfVersion.LessThan(runner.tfBinContext.DefaultTfVersion) {
			if runner.tfBinContext.TfUpgradePath == nil || len(runner.tfBinContext.TfUpgradePath) == 0 {
				return errors.New("terraform version mismatch and no upgrade path specified")
			}
			for _, targetTfVersion := range runner.tfBinContext.TfUpgradePath {
				if currentTfVersion.LessThan(targetTfVersion.GetTerraformVersion()) {
					err = workspace.Apply(ctx, runner.VersionedExecutor(targetTfVersion.GetTerraformVersion()))
					if err != nil {
						return err
					}
				}
			}
		}
	} else if currentTfVersion.LessThan(runner.tfBinContext.DefaultTfVersion) {
		return errors.New("apply attempted with a newer version of terraform than the state")
	}

	return nil
}

// Destroy runs `terraform destroy` on the given workspace in the background.
// The status of the job can be found by polling the Status function.
func (runner *TfJobRunner) Destroy(ctx context.Context, id string, templateVars map[string]interface{}) error {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return err
	}

	workspace, err := wrapper.DeserializeWorkspace(deployment.Workspace)
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
		err := workspace.Destroy(ctx, runner.DefaultExecutor())
		runner.operationFinished(err, workspace, deployment)
	}()

	return nil
}

// operationFinished closes out the state of the background job so clients that
// are polling can get the results.
func (runner *TfJobRunner) operationFinished(err error, workspace Workspace, deployment storage.TerraformDeployment) error {
	// we shouldn't update the status on update when updating the HCL, as the status comes either from the provision call or a previous update
	if err == nil {
		lastOperationMessage := ""
		// maybe do if deployment.LastOperationType != "validation" so we don't do the status update on staging a job.
		// previously we would only stage a job on provision so state would be empty and the outputs would be null.
		outputs, err := workspace.Outputs(workspace.ModuleInstances()[0].InstanceName)
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

	ws, err := wrapper.DeserializeWorkspace(deployment.Workspace)
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

// Show returns the output from terraform show command
func (runner *TfJobRunner) Show(ctx context.Context, id string) (string, error) {
	deployment, err := runner.store.GetTerraformDeployment(id)
	if err != nil {
		return "", err
	}

	workspace, err := wrapper.DeserializeWorkspace(deployment.Workspace)
	if err != nil {
		return "", err
	}

	return workspace.Show(ctx, runner.DefaultExecutor())
}
