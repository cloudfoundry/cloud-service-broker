package tf

import (
	"github.gwd.broadcom.net/TNZ/brokerapi/v13/domain/apiresponses"
	"github.gwd.broadcom.net/TNZ/cloud-service-broker/v2/dbservice/models"
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
