package tf

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/featureflags"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
)

type DeploymentManager struct {
	store broker.ServiceProviderStorage
}

func NewDeploymentManager(store broker.ServiceProviderStorage) *DeploymentManager {
	return &DeploymentManager{
		store: store,
	}
}

func (d *DeploymentManager) CreateAndSaveDeployment(deploymentID string, workspace *workspace.TerraformWorkspace) (storage.TerraformDeployment, error) {
	deployment := storage.TerraformDeployment{ID: deploymentID}
	exists, err := d.store.ExistsTerraformDeployment(deploymentID)
	switch {
	case err != nil:
		return deployment, err
	case exists:
		deployment, err = d.store.GetTerraformDeployment(deploymentID)
		if err != nil {
			return deployment, err
		}
	}

	deployment.Workspace = workspace

	return deployment, d.store.StoreTerraformDeployment(deployment)
}

func (d *DeploymentManager) MarkOperationStarted(deployment *storage.TerraformDeployment, operationType string) error {
	deployment.LastOperationType = operationType
	deployment.LastOperationState = InProgress
	deployment.LastOperationMessage = fmt.Sprintf("%s %s", operationType, InProgress)

	if err := d.store.StoreTerraformDeployment(*deployment); err != nil {
		return err
	}

	return nil
}

func (d *DeploymentManager) MarkOperationFinished(deployment *storage.TerraformDeployment, err error) error {
	if err == nil {
		lastOperationMessage := fmt.Sprintf("%s %s", deployment.LastOperationType, Succeeded)
		outputs, err := deployment.Workspace.Outputs(deployment.Workspace.ModuleInstances()[0].InstanceName)
		if err == nil {
			if status, ok := outputs["status"]; ok {
				lastOperationMessage = fmt.Sprintf("%s %s: %s", deployment.LastOperationType, Succeeded, status)
			}
		}
		deployment.LastOperationState = Succeeded
		deployment.LastOperationMessage = lastOperationMessage
	} else {
		deployment.LastOperationState = Failed
		deployment.LastOperationMessage = fmt.Sprintf("%s %s: %s", deployment.LastOperationType, Failed, err)
	}

	return d.store.StoreTerraformDeployment(*deployment)
}

func (d *DeploymentManager) OperationStatus(deploymentID string) (bool, string, error) {
	deployment, err := d.store.GetTerraformDeployment(deploymentID)
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

func (d *DeploymentManager) UpdateWorkspaceHCL(deploymentID string, serviceDefinitionAction TfServiceDefinitionV1Action, templateVars map[string]any) error {
	if !featureflags.Enabled(featureflags.DynamicHCLEnabled) && !featureflags.Enabled(featureflags.TfUpgradeEnabled) {
		return nil
	}
	deployment, err := d.store.GetTerraformDeployment(deploymentID)
	if err != nil {
		return err
	}

	currentWorkspace := deployment.TFWorkspace()
	if err != nil {
		return err
	}

	newWorkspace, err := workspace.NewWorkspace(templateVars, serviceDefinitionAction.Template, serviceDefinitionAction.Templates, []workspace.ParameterMapping{}, []string{}, []workspace.ParameterMapping{})
	if err != nil {
		return err
	}

	newWorkspace.State = currentWorkspace.State

	deployment.Workspace = newWorkspace
	if err := d.store.StoreTerraformDeployment(deployment); err != nil {
		return fmt.Errorf("terraform provider create failed: %w", err)
	}

	return nil
}

func (d *DeploymentManager) GetTerraformDeployment(deploymentID string) (storage.TerraformDeployment, error) {
	return d.store.GetTerraformDeployment(deploymentID)
}

func (d *DeploymentManager) GetBindingDeployments(deploymentID string) ([]storage.TerraformDeployment, error) {
	instanceID := getInstanceIDFromTfID(deploymentID)
	bindingIDs, err := d.store.GetServiceBindingIDsForServiceInstance(instanceID)
	if err != nil {
		return []storage.TerraformDeployment{}, err
	}

	var bindingDeployments []storage.TerraformDeployment
	for _, bindingID := range bindingIDs {
		bindingDeployment, err := d.store.GetTerraformDeployment(generateTfID(instanceID, bindingID))
		if err != nil {
			return []storage.TerraformDeployment{}, err
		}

		bindingDeployments = append(bindingDeployments, bindingDeployment)
	}
	return bindingDeployments, nil
}
