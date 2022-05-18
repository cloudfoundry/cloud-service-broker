package tf

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
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

func (d *DeploymentManager) CreateAndSaveDeployment(jobID string, workspace *workspace.TerraformWorkspace) (storage.TerraformDeployment, error) {
	deployment := storage.TerraformDeployment{ID: jobID}
	exists, err := d.store.ExistsTerraformDeployment(jobID)
	switch {
	case err != nil:
		return deployment, err
	case exists:
		deployment, err = d.store.GetTerraformDeployment(jobID)
		if err != nil {
			return deployment, err
		}
	}

	deployment.Workspace = workspace
	deployment.LastOperationType = "validation"

	return deployment, d.store.StoreTerraformDeployment(deployment)
}

func (d *DeploymentManager) MarkJobStarted(deployment storage.TerraformDeployment, operationType string) error {
	// update the deployment info
	deployment.LastOperationType = operationType
	deployment.LastOperationState = InProgress
	deployment.LastOperationMessage = ""

	if err := d.store.StoreTerraformDeployment(deployment); err != nil {
		return err
	}

	return nil
}

// OperationFinished closes out the state of the background job so clients that
// are polling can get the results.
func (d *DeploymentManager) OperationFinished(err error, deployment storage.TerraformDeployment) error {
	// we shouldn't update the Status on update when updating the HCL, as the Status comes either from the provision call or a previous update
	workspace := deployment.Workspace
	if err == nil {
		lastOperationMessage := ""
		// maybe do if deployment.LastOperationType != "validation" so we don't do the Status update on staging a job.
		// previously we would only stage a job on provision so state would be empty and the outputs would be null.
		outputs, err := workspace.Outputs(workspace.ModuleInstances()[0].InstanceName)
		if err == nil {
			if status, ok := outputs["Status"]; ok {
				lastOperationMessage = fmt.Sprintf("%v", status)
			}
		}
		deployment.LastOperationState = Succeeded
		deployment.LastOperationMessage = lastOperationMessage
	} else {
		deployment.LastOperationState = Failed
		deployment.LastOperationMessage = err.Error()
	}

	return d.store.StoreTerraformDeployment(deployment)
}

func (d *DeploymentManager) Status(deploymentID string) (bool, string, error) {
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

func (d *DeploymentManager) GetTerraformDeployment(id string) (storage.TerraformDeployment, error) {
	return d.store.GetTerraformDeployment(id)
}

func (d *DeploymentManager) UpdateWorkspaceHCL(action TfServiceDefinitionV1Action, operationContext *varcontext.VarContext, tfID string) error {
	if !viper.GetBool(dynamicHCLEnabled) {
		return nil
	}
	deployment, err := d.store.GetTerraformDeployment(tfID)
	if err != nil {
		return err
	}

	currentWorkspace := deployment.TFWorkspace()
	if err != nil {
		return err
	}

	workspace, err := workspace.NewWorkspace(operationContext.ToMap(), action.Template, action.Templates, []workspace.ParameterMapping{}, []string{}, []workspace.ParameterMapping{})
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
