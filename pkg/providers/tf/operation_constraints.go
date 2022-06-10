package tf

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
)

func (provider *TerraformProvider) CheckUpgradeAvailable(deploymentGUID string) error {
	deployment, err := provider.GetTerraformDeployment(deploymentGUID)
	if err != nil {
		return err
	}

	workspace := deployment.TFWorkspace()

	err = provider.checkTerraformVersion(workspace)
	if err != nil {
		return err
	}

	return nil
}

func (provider *TerraformProvider) checkTerraformVersion(workspace workspace.Workspace) error {
	currentTfVersion, err := workspace.StateTFVersion()
	if err != nil {
		return err
	}
	if currentTfVersion.LessThan(provider.tfBinContext.DefaultTfVersion) {
		return errors.New("operation attempted with newer version of Terraform than current state, upgrade the service before retrying operation")
	}
	return nil
}

func (provider *TerraformProvider) CheckOperationConstraints(deploymentID string, operationType string) error {
	if operationType != models.DeprovisionOperationType {
		return nil
	}

	deployment, err := provider.GetTerraformDeployment(deploymentID)
	if err != nil {
		return err
	}

	isProvisionOperationInProgress := deployment.LastOperationType == models.ProvisionOperationType && deployment.LastOperationState == InProgress
	if isProvisionOperationInProgress {
		return fmt.Errorf("destroy operation not allowed while provision is in progress")
	}

	return nil
}
