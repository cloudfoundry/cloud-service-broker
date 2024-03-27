package dbservice

import (
	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"gorm.io/gorm"
)

func recoverInProgressOperations(db *gorm.DB, logger lager.Logger) error {
	logger = logger.Session("recover-in-progress-operations")

	var terraformDeploymentBatch []models.TerraformDeployment
	result := db.Where("last_operation_state = ?", "in progress").FindInBatches(&terraformDeploymentBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range terraformDeploymentBatch {
			terraformDeploymentBatch[i].LastOperationState = "failed"
			terraformDeploymentBatch[i].LastOperationMessage = "the broker restarted while the operation was in progress"
			logger.Info("mark-as-failed", lager.Data{"workspace_id": terraformDeploymentBatch[i].ID})
		}

		return tx.Save(&terraformDeploymentBatch).Error
	})

	return result.Error
}
