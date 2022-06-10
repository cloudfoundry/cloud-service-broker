package tf

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
)

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
		return fmt.Errorf("operation not allowed while provision is in progress")
	}

	return nil
}
