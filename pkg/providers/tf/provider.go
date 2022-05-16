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

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/hclparser"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . JobRunner
type JobRunner interface {
	Import(ctx context.Context, id string, importResources []ImportResource) error
	Create(ctx context.Context, id string) error
	Update(ctx context.Context, id string, templateVars map[string]interface{}) error
	Destroy(ctx context.Context, id string, templateVars map[string]interface{}) error
	Status(ctx context.Context, id string) (bool, string, error)
	Outputs(ctx context.Context, id, instanceName string) (map[string]interface{}, error)
	Wait(ctx context.Context, id string) error
	Show(ctx context.Context, id string) (string, error)
}

// NewTerraformProvider creates a new ServiceProvider backed by Terraform module definitions for provision and bind.
func NewTerraformProvider(jobRunner JobRunner, logger lager.Logger, serviceDefinition TfServiceDefinitionV1, store broker.ServiceProviderStorage) broker.ServiceProvider {
	return &terraformProvider{
		serviceDefinition: serviceDefinition,
		jobRunner:         jobRunner,
		logger:            logger.Session("terraform-" + serviceDefinition.Name),
		store:             store,
	}
}

type terraformProvider struct {
	logger            lager.Logger
	jobRunner         JobRunner
	serviceDefinition TfServiceDefinitionV1
	store             broker.ServiceProviderStorage
}

// BuildInstanceCredentials combines the bind credentials with the connection
// information in the instance details to get a full set of connection details.
func (provider *terraformProvider) BuildInstanceCredentials(ctx context.Context, credentials map[string]interface{}, outputs storage.JSONObject) (*domain.Binding, error) {
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
func (provider *terraformProvider) Provision(ctx context.Context, provisionContext *varcontext.VarContext) (storage.ServiceInstanceDetails, error) {
	provider.logger.Debug("terraform-provision", correlation.ID(ctx), lager.Data{
		"context": provisionContext.ToMap(),
	})

	var (
		tfID string
		err  error
	)

	switch provider.serviceDefinition.ProvisionSettings.IsTfImport(provisionContext) {
	case true:
		tfID, err = provider.importCreate(ctx, provisionContext, provider.serviceDefinition.ProvisionSettings)
	default:
		tfID, err = provider.create(ctx, provisionContext, provider.serviceDefinition.ProvisionSettings)
	}
	if err != nil {
		return storage.ServiceInstanceDetails{}, err
	}

	return storage.ServiceInstanceDetails{
		OperationGUID: tfID,
		OperationType: models.ProvisionOperationType,
	}, nil
}

// Update makes necessary updates to resources so they match new desired configuration
func (provider *terraformProvider) Update(ctx context.Context, provisionContext *varcontext.VarContext) (models.ServiceInstanceDetails, error) {
	provider.logger.Debug("update", correlation.ID(ctx), lager.Data{
		"context": provisionContext.ToMap(),
	})

	if provider.serviceDefinition.ProvisionSettings.IsTfImport(provisionContext) {
		return models.ServiceInstanceDetails{}, fmt.Errorf("cannot update to subsume plan\n\nFor OpsMan Tile users see documentation here: https://via.vmw.com/ENs4\n\nFor Open Source users deployed via 'cf push' see documentation here:  https://via.vmw.com/ENw4")
	}

	tfID := provisionContext.GetString("tf_id")
	if err := provisionContext.Error(); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	if err := UpdateWorkspaceHCL(provider.store, provider.serviceDefinition.ProvisionSettings, provisionContext, tfID); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	err := provider.jobRunner.Update(ctx, tfID, provisionContext.ToMap())

	return models.ServiceInstanceDetails{
		OperationID:   tfID,
		OperationType: models.UpdateOperationType,
	}, err
}

// Bind creates a new backing Terraform job and executes it, waiting on the result.
func (provider *terraformProvider) Bind(ctx context.Context, bindContext *varcontext.VarContext) (map[string]interface{}, error) {
	provider.logger.Debug("terraform-bind", correlation.ID(ctx), lager.Data{
		"context": bindContext.ToMap(),
	})

	tfID, err := provider.create(ctx, bindContext, provider.serviceDefinition.BindSettings)
	if err != nil {
		return nil, fmt.Errorf("error from provider bind: %w", err)
	}

	if err := provider.jobRunner.Wait(ctx, tfID); err != nil {
		return nil, fmt.Errorf("error from job runner: %w", err)
	}

	return provider.jobRunner.Outputs(ctx, tfID, workspace.DefaultInstanceName)
}

func (provider *terraformProvider) importCreate(ctx context.Context, vars *varcontext.VarContext, action TfServiceDefinitionV1Action) (string, error) {
	varsMap := vars.ToMap()

	var parameterMappings, addParams []workspace.ParameterMapping

	for _, importParameterMapping := range action.ImportParameterMappings {
		parameterMappings = append(parameterMappings, workspace.ParameterMapping{
			TfVariable:    importParameterMapping.TfVariable,
			ParameterName: importParameterMapping.ParameterName,
		})
	}

	for _, addParam := range action.ImportParametersToAdd {
		addParams = append(addParams, workspace.ParameterMapping{
			TfVariable:    addParam.TfVariable,
			ParameterName: addParam.ParameterName,
		})
	}

	var importParams []ImportResource

	for _, importParam := range action.ImportVariables {
		if param, ok := varsMap[importParam.Name]; ok {
			importParams = append(importParams, ImportResource{TfResource: importParam.TfResource, IaaSResource: fmt.Sprintf("%v", param)})
		}
	}

	if len(importParams) != len(action.ImportVariables) {
		importFields := action.ImportVariables[0].Name
		for i := 1; i < len(action.ImportVariables); i++ {
			importFields = fmt.Sprintf("%s, %s", importFields, action.ImportVariables[i].Name)
		}

		return "", fmt.Errorf("must provide values for all import parameters: %s", importFields)
	}

	tfID := vars.GetString("tf_id")
	if err := vars.Error(); err != nil {
		return "", err
	}

	workspace, err := workspace.NewWorkspace(varsMap, "", action.Templates, parameterMappings, action.ImportParametersToDelete, addParams)
	if err != nil {
		return tfID, err
	}

	if err := provider.storeDeployment(tfID, workspace); err != nil {
		provider.logger.Error("terraform provider create failed", err)
		return tfID, err
	}

	return tfID, provider.jobRunner.Import(ctx, tfID, importParams)
}

func (provider *terraformProvider) create(ctx context.Context, vars *varcontext.VarContext, action TfServiceDefinitionV1Action) (string, error) {
	tfID := vars.GetString("tf_id")
	if err := vars.Error(); err != nil {
		return "", err
	}

	workspace, err := workspace.NewWorkspace(vars.ToMap(), action.Template, action.Templates, []workspace.ParameterMapping{}, []string{}, []workspace.ParameterMapping{})
	if err != nil {
		return tfID, fmt.Errorf("error creating workspace: %w", err)
	}

	// if err = workspace.Validate(); err != nil {
	// 	return tfID, err
	// }

	if err := provider.storeDeployment(tfID, workspace); err != nil {
		provider.logger.Error("terraform provider create failed", err)
		return tfID, fmt.Errorf("terraform provider create failed: %w", err)
	}

	return tfID, provider.jobRunner.Create(ctx, tfID)
}

// storeDeployment stages a job to be executed. Before the workspace is saved to the
// database, the modules and inputs are validated by Terraform.
func (provider *terraformProvider) storeDeployment(jobID string, workspace *workspace.TerraformWorkspace) error {
	deployment := storage.TerraformDeployment{ID: jobID}
	exists, err := provider.store.ExistsTerraformDeployment(jobID)
	switch {
	case err != nil:
		return err
	case exists:
		deployment, err = provider.store.GetTerraformDeployment(jobID)
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

	return provider.store.StoreTerraformDeployment(deployment)
}

// Unbind performs a terraform destroy on the binding.
func (provider *terraformProvider) Unbind(ctx context.Context, instanceGUID, bindingID string, vc *varcontext.VarContext) error {
	tfID := generateTfID(instanceGUID, bindingID)
	provider.logger.Debug("terraform-unbind", correlation.ID(ctx), lager.Data{
		"instance": instanceGUID,
		"binding":  bindingID,
		"tfId":     tfID,
	})

	if err := UpdateWorkspaceHCL(provider.store, provider.serviceDefinition.BindSettings, vc, tfID); err != nil {
		return err
	}

	if err := provider.jobRunner.Destroy(ctx, tfID, vc.ToMap()); err != nil {
		return err
	}

	return provider.jobRunner.Wait(ctx, tfID)
}

// Deprovision performs a terraform destroy on the instance.
func (provider *terraformProvider) Deprovision(ctx context.Context, instanceGUID string, details domain.DeprovisionDetails, vc *varcontext.VarContext) (operationID *string, err error) {
	provider.logger.Debug("terraform-deprovision", correlation.ID(ctx), lager.Data{
		"instance": instanceGUID,
	})

	tfID := generateTfID(instanceGUID, "")

	if err := UpdateWorkspaceHCL(provider.store, provider.serviceDefinition.ProvisionSettings, vc, tfID); err != nil {
		return nil, err
	}

	if err := provider.jobRunner.Destroy(ctx, tfID, vc.ToMap()); err != nil {
		return nil, err
	}

	return &tfID, nil
}

// PollInstance returns the instance status of the backing job.
func (provider *terraformProvider) PollInstance(ctx context.Context, instanceGUID string) (bool, string, error) {
	return provider.jobRunner.Status(ctx, generateTfID(instanceGUID, ""))
}

// UpdateInstanceDetails updates the ServiceInstanceDetails with the most recent state from GCP.
// This function is optional, but will be called after async provisions, updates, and possibly
// on broker version changes.
// Return a nil error if you choose not to implement this function.
func (provider *terraformProvider) GetTerraformOutputs(ctx context.Context, guid string) (storage.JSONObject, error) {
	tfID := generateTfID(guid, "")

	outs, err := provider.jobRunner.Outputs(ctx, tfID, workspace.DefaultInstanceName)
	if err != nil {
		return nil, err
	}

	return outs, nil
}

func (provider *terraformProvider) GetImportedProperties(ctx context.Context, planGUID string, instanceGUID string, inputVariables []broker.BrokerVariable) (map[string]interface{}, error) {
	provider.logger.Debug("getImportedProperties", correlation.ID(ctx), lager.Data{})

	if provider.isSubsumePlan(planGUID) {
		return map[string]interface{}{}, nil
	}

	varsToReplace := provider.getVarsToReplace(inputVariables)
	if len(varsToReplace) == 0 {
		return map[string]interface{}{}, nil
	}

	tfHCL, err := provider.jobRunner.Show(ctx, generateTfID(instanceGUID, ""))
	if err != nil {
		return map[string]interface{}{}, err
	}

	return hclparser.GetParameters(tfHCL, varsToReplace)
}

func (provider *terraformProvider) getVarsToReplace(inputVariables []broker.BrokerVariable) []hclparser.ExtractVariable {
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

func (provider *terraformProvider) isSubsumePlan(planGUID string) bool {
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
