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

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/invoker"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"

	"github.com/hashicorp/go-version"

	"github.com/spf13/viper"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
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

// ImportResource represents TF resource to IaaS resource ID mapping for import
type ImportResource struct {
	TfResource   string
	IaaSResource string
}

func (provider *TerraformProvider) terraformPlanToCheckNoResourcesDeleted(invoker invoker.TerraformInvoker, ctx context.Context, workspace *workspace.TerraformWorkspace, logger lager.Logger) error {
	planOutput, err := invoker.Plan(ctx, workspace)
	if err != nil {
		return err
	}
	err = CheckTerraformPlanOutput(logger, planOutput)
	return err
}

func (provider *TerraformProvider) DefaultInvoker() invoker.TerraformInvoker {
	return provider.VersionedInvoker(provider.tfBinContext.DefaultTfVersion)
}

func (provider *TerraformProvider) VersionedInvoker(version *version.Version) invoker.TerraformInvoker {
	return provider.VersionedTerraformInvoker(version)
}

func (provider *TerraformProvider) performTerraformUpgrade(ctx context.Context, workspace workspace.Workspace) error {
	currentTfVersion, err := workspace.StateVersion()
	if err != nil {
		return err
	}

	if viper.GetBool(TfUpgradeEnabled) {
		if currentTfVersion.LessThan(provider.tfBinContext.DefaultTfVersion) {
			if provider.tfBinContext.TfUpgradePath == nil || len(provider.tfBinContext.TfUpgradePath) == 0 {
				return errors.New("terraform version mismatch and no upgrade path specified")
			}
			for _, targetTfVersion := range provider.tfBinContext.TfUpgradePath {
				if currentTfVersion.LessThan(targetTfVersion) {
					err = provider.VersionedInvoker(targetTfVersion).Apply(ctx, workspace)
					if err != nil {
						return err
					}
				}
			}
		}
	} else if currentTfVersion.LessThan(provider.tfBinContext.DefaultTfVersion) {
		return errors.New("apply attempted with a newer version of terraform than the state")
	}

	return nil
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

// Outputs gets the output variables for the given module instance in the workspace.
func (provider *TerraformProvider) Outputs(ctx context.Context, id, instanceName string) (map[string]interface{}, error) {
	deployment, err := provider.store.GetTerraformDeployment(id)
	if err != nil {
		return nil, fmt.Errorf("error getting TF deployment: %w", err)
	}

	return deployment.Workspace.Outputs(instanceName)
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
