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

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/invoker"
	"github.com/hashicorp/go-version"
	"github.com/spf13/viper"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/hclparser"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
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

// NewTerraformProvider creates a new ServiceProvider backed by Terraform module definitions for provision and bind.
func NewTerraformProvider(
	tfBinContext executor.TFBinariesContext,
	invokerBuilder invoker.TerraformInvokerBuilder,
	logger lager.Logger,
	serviceDefinition TfServiceDefinitionV1,
	store broker.ServiceProviderStorage,
) *TerraformProvider {
	return &TerraformProvider{
		tfBinContext:            tfBinContext,
		TerraformInvokerBuilder: invokerBuilder,
		serviceDefinition:       serviceDefinition,
		logger:                  logger.Session("terraform-" + serviceDefinition.Name),
		store:                   store,
	}
}

type TerraformProvider struct {
	tfBinContext executor.TFBinariesContext
	invoker.TerraformInvokerBuilder
	logger            lager.Logger
	serviceDefinition TfServiceDefinitionV1
	store             broker.ServiceProviderStorage
}

func (provider *TerraformProvider) DefaultInvoker() invoker.TerraformInvoker {
	return provider.VersionedInvoker(provider.tfBinContext.DefaultTfVersion)
}

func (provider *TerraformProvider) VersionedInvoker(version *version.Version) invoker.TerraformInvoker {
	return provider.VersionedTerraformInvoker(version)
}

func (provider *TerraformProvider) create(ctx context.Context, vars *varcontext.VarContext, action TfServiceDefinitionV1Action) (string, error) {
	tfID := vars.GetString("tf_id")
	if err := vars.Error(); err != nil {
		return "", err
	}

	workspace, err := workspace.NewWorkspace(vars.ToMap(), action.Template, action.Templates, []workspace.ParameterMapping{}, []string{}, []workspace.ParameterMapping{})
	if err != nil {
		return tfID, fmt.Errorf("error creating workspace: %w", err)
	}

	deployment, err := provider.createAndSaveDeployment(tfID, workspace)
	if err != nil {
		provider.logger.Error("terraform provider create failed", err)
		return tfID, fmt.Errorf("terraform provider create failed: %w", err)
	}

	if err := provider.markJobStarted(deployment, models.ProvisionOperationType); err != nil {
		return tfID, fmt.Errorf("error marking job started: %w", err)
	}

	go func() {
		err := provider.DefaultInvoker().Apply(ctx, workspace)
		provider.operationFinished(err, deployment)
	}()

	return tfID, nil
}

// Destroy runs `terraform destroy` on the given workspace in the background.
// The status of the job can be found by polling the Status function.
func (provider *TerraformProvider) Destroy(ctx context.Context, id string, templateVars map[string]interface{}) error {
	deployment, err := provider.store.GetTerraformDeployment(id)
	if err != nil {
		return err
	}

	workspace := deployment.TFWorkspace()

	inputList, err := workspace.Modules[0].Inputs()
	if err != nil {
		return err
	}

	limitedConfig := make(map[string]interface{})
	for _, name := range inputList {
		limitedConfig[name] = templateVars[name]
	}

	workspace.Instances[0].Configuration = limitedConfig

	if err := provider.markJobStarted(deployment, models.DeprovisionOperationType); err != nil {
		return err
	}

	go func() {
		err = provider.performTerraformUpgrade(ctx, workspace)
		if err != nil {
			provider.operationFinished(err, deployment)
			return
		}

		err = provider.DefaultInvoker().Destroy(ctx, workspace)
		provider.operationFinished(err, deployment)
	}()

	return nil
}

func (provider *TerraformProvider) createAndSaveDeployment(jobID string, workspace *workspace.TerraformWorkspace) (storage.TerraformDeployment, error) {
	deployment := storage.TerraformDeployment{ID: jobID}
	exists, err := provider.store.ExistsTerraformDeployment(jobID)
	switch {
	case err != nil:
		return deployment, err
	case exists:
		deployment, err = provider.store.GetTerraformDeployment(jobID)
		if err != nil {
			return deployment, err
		}
	}

	deployment.Workspace = workspace
	deployment.LastOperationType = "validation"

	return deployment, provider.store.StoreTerraformDeployment(deployment)
}

func (provider *TerraformProvider) GetTerraformOutputs(ctx context.Context, guid string) (storage.JSONObject, error) {
	tfID := generateTfID(guid, "")

	outs, err := provider.Outputs(ctx, tfID, workspace.DefaultInstanceName)
	if err != nil {
		return nil, err
	}

	return outs, nil
}

func (provider *TerraformProvider) GetImportedProperties(ctx context.Context, planGUID string, instanceGUID string, inputVariables []broker.BrokerVariable) (map[string]interface{}, error) {
	provider.logger.Debug("getImportedProperties", correlation.ID(ctx), lager.Data{})

	if provider.serviceDefinition.IsSubsumePlan(planGUID) {
		return map[string]interface{}{}, nil
	}

	varsToReplace := provider.getVarsToReplace(inputVariables)
	if len(varsToReplace) == 0 {
		return map[string]interface{}{}, nil
	}

	deployment, err := provider.store.GetTerraformDeployment(generateTfID(instanceGUID, ""))
	if err != nil {
		return nil, err
	}

	tfHCL, err := provider.DefaultInvoker().Show(ctx, deployment.Workspace)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return hclparser.GetParameters(tfHCL, varsToReplace)
}

func (provider *TerraformProvider) getVarsToReplace(inputVariables []broker.BrokerVariable) []hclparser.ExtractVariable {
	var varsToReplace []hclparser.ExtractVariable
	for _, vars := range inputVariables {
		if vars.TFAttribute != "" {
			varsToReplace = append(varsToReplace, hclparser.ExtractVariable{
				FieldToRead:  vars.TFAttribute,
				FieldToWrite: vars.FieldName,
			})
		}
	}
	return varsToReplace
}

func (provider *TerraformProvider) markJobStarted(deployment storage.TerraformDeployment, operationType string) error {
	// update the deployment info
	deployment.LastOperationType = operationType
	deployment.LastOperationState = InProgress
	deployment.LastOperationMessage = ""

	if err := provider.store.StoreTerraformDeployment(deployment); err != nil {
		return err
	}

	return nil
}

// operationFinished closes out the state of the background job so clients that
// are polling can get the results.
func (provider *TerraformProvider) operationFinished(err error, deployment storage.TerraformDeployment) error {
	// we shouldn't update the status on update when updating the HCL, as the status comes either from the provision call or a previous update
	workspace := deployment.Workspace
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

	return provider.store.StoreTerraformDeployment(deployment)
}

// Status gets the status of the most recent job on the workspace.
// If isDone is true, then the status of the operation will not change again.
// if isDone is false, then the operation is ongoing.
func (provider *TerraformProvider) Status(ctx context.Context, id string) (bool, string, error) {
	deployment, err := provider.store.GetTerraformDeployment(id)
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

// Wait waits for an operation to complete, polling its status once per second.
func (provider *TerraformProvider) Wait(ctx context.Context, id string) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		case <-time.After(1 * time.Second):
			isDone, _, err := provider.Status(ctx, id)
			if isDone {
				return err
			}
		}
	}
}

// Outputs gets the output variables for the given module instance in the workspace.
func (provider *TerraformProvider) Outputs(ctx context.Context, id, instanceName string) (map[string]interface{}, error) {
	deployment, err := provider.store.GetTerraformDeployment(id)
	if err != nil {
		return nil, fmt.Errorf("error getting TF deployment: %w", err)
	}

	return deployment.Workspace.Outputs(instanceName)
}
