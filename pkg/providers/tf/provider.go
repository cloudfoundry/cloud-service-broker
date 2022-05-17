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
	"fmt"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/invoker"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/hclparser"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

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

// BuildInstanceCredentials combines the bind credentials with the connection
// information in the instance details to get a full set of connection details.
func (provider *TerraformProvider) BuildInstanceCredentials(ctx context.Context, credentials map[string]interface{}, outputs storage.JSONObject) (*domain.Binding, error) {
	vc, err := varcontext.Builder().
		MergeMap(outputs).
		MergeMap(credentials).
		Build()
	if err != nil {
		return nil, err
	}

	return &domain.Binding{Credentials: vc.ToMap()}, nil
}

// Provision creates the necessary resources that an instance of this service
// needs to operate.

// Update makes necessary updates to resources so they match new desired configuration

// Bind creates a new backing Terraform job and executes it, waiting on the result.
func (provider *TerraformProvider) Bind(ctx context.Context, bindContext *varcontext.VarContext) (map[string]interface{}, error) {
	provider.logger.Debug("terraform-bind", correlation.ID(ctx), lager.Data{
		"context": bindContext.ToMap(),
	})

	tfID, err := provider.create(ctx, bindContext, provider.serviceDefinition.BindSettings)
	if err != nil {
		return nil, fmt.Errorf("error from provider bind: %w", err)
	}

	if err := provider.Wait(ctx, tfID); err != nil {
		return nil, fmt.Errorf("error from job runner: %w", err)
	}

	return provider.Outputs(ctx, tfID, workspace.DefaultInstanceName)
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

// Unbind performs a terraform destroy on the binding.
func (provider *TerraformProvider) Unbind(ctx context.Context, instanceGUID, bindingID string, vc *varcontext.VarContext) error {
	tfID := generateTfID(instanceGUID, bindingID)
	provider.logger.Debug("terraform-unbind", correlation.ID(ctx), lager.Data{
		"instance": instanceGUID,
		"binding":  bindingID,
		"tfId":     tfID,
	})

	if err := UpdateWorkspaceHCL(provider.store, provider.serviceDefinition.BindSettings, vc, tfID); err != nil {
		return err
	}

	if err := provider.Destroy(ctx, tfID, vc.ToMap()); err != nil {
		return err
	}

	return provider.Wait(ctx, tfID)
}

// Deprovision performs a terraform destroy on the instance.
func (provider *TerraformProvider) Deprovision(ctx context.Context, instanceGUID string, details domain.DeprovisionDetails, vc *varcontext.VarContext) (operationID *string, err error) {
	provider.logger.Debug("terraform-deprovision", correlation.ID(ctx), lager.Data{
		"instance": instanceGUID,
	})

	tfID := generateTfID(instanceGUID, "")

	if err := UpdateWorkspaceHCL(provider.store, provider.serviceDefinition.ProvisionSettings, vc, tfID); err != nil {
		return nil, err
	}

	if err := provider.Destroy(ctx, tfID, vc.ToMap()); err != nil {
		return nil, err
	}

	return &tfID, nil
}

// PollInstance returns the instance status of the backing job.
func (provider *TerraformProvider) PollInstance(ctx context.Context, instanceGUID string) (bool, string, error) {
	return provider.Status(ctx, generateTfID(instanceGUID, ""))
}

// UpdateInstanceDetails updates the ServiceInstanceDetails with the most recent state from GCP.
// This function is optional, but will be called after async provisions, updates, and possibly
// on broker version changes.
// Return a nil error if you choose not to implement this function.
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

	if provider.isSubsumePlan(planGUID) {
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

func (provider *TerraformProvider) isSubsumePlan(planGUID string) bool {
	for _, plan := range provider.serviceDefinition.Plans {
		if plan.ID == planGUID {
			if _, ok := plan.Properties["subsume"]; !ok {
				return true
			}
			break
		}
	}
	return false
}
