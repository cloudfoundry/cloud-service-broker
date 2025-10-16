package tf

import (
	"code.cloudfoundry.org/brokerapi/v13/domain/apiresponses"
	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
)

func (provider *TerraformProvider) CheckOperationConstraints(deploymentID string, operationType string) error {
	if operationType != models.DeprovisionOperationType {
		return nil
	}

	deployment, err := provider.GetTerraformDeployment(deploymentID)
	switch {
	case err != nil:
		return err
	case deployment.LastOperationType == models.ProvisionOperationType && deployment.LastOperationState == InProgress:
		// Will not accept a deprovision while a provision is in progress
		return apiresponses.ErrConcurrentInstanceAccess
	default:
		return nil
	}
}
