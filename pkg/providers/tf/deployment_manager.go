package tf

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/spf13/viper"
)

type DeploymentManager struct {
	store broker.ServiceProviderStorage
}

func NewDeploymentManager(store broker.ServiceProviderStorage) *DeploymentManager {
	return &DeploymentManager{
		store: store,
	}
}

const (
	dynamicHCLEnabled = "brokerpak.updates.enabled"
)

func init() {
	viper.BindEnv(dynamicHCLEnabled, "BROKERPAK_UPDATES_ENABLED")
	viper.SetDefault(dynamicHCLEnabled, false)
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
		workspace := deployment.Workspace
		outputs, err := workspace.Outputs(workspace.ModuleInstances()[0].InstanceName)
		if err == nil {
			if status, ok := outputs["status"]; ok {
				lastOperationMessage = fmt.Sprintf("%s %s: %v", deployment.LastOperationType, Succeeded, status)
			}
		}
		deployment.LastOperationState = Succeeded
		deployment.LastOperationMessage = lastOperationMessage
	} else {
		deployment.LastOperationState = Failed
		deployment.LastOperationMessage = fmt.Errorf("%s %s: %w", deployment.LastOperationType, Failed, err).Error()
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

func (d *DeploymentManager) UpdateWorkspaceHCL(deploymentID string, serviceDefinitionAction TfServiceDefinitionV1Action, templateVars map[string]interface{}) error {
	if !viper.GetBool(dynamicHCLEnabled) {
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

	workspace, err := workspace.NewWorkspace(templateVars, serviceDefinitionAction.Template, serviceDefinitionAction.Templates, []workspace.ParameterMapping{}, []string{}, []workspace.ParameterMapping{})
	if err != nil {
		return err
	}

	workspace.State = currentWorkspace.State

	deployment.Workspace = workspace
	if err := d.store.StoreTerraformDeployment(deployment); err != nil {
		return fmt.Errorf("terraform provider create failed: %w", err)
	}

	return nil
}

func (d *DeploymentManager) GetTerraformDeployment(deploymentID string) (storage.TerraformDeployment, error) {
	return d.store.GetTerraformDeployment(deploymentID)
}
