package storage

import (
	"os"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"gorm.io/gorm"
)

func (s *Storage) RecoverInProgressOperations(logger lager.Logger) error {
	logger = logger.Session("recover-in-progress-operations")

	// We only wan't to fail interrupted service instances if we detect that we run as a CF APP.
	// VM based csb instances implement a drain mechanism and should need this. Additionally VM
	// based csb deployments are scalable horizontally and the below would fail in flight instances
	// of another csb process.
	if os.Getenv("CF_INSTANCE_GUID") != "" {
		var terraformDeploymentBatch []models.TerraformDeployment
		result := s.db.Where("last_operation_state = ?", "in progress").FindInBatches(&terraformDeploymentBatch, 100, func(tx *gorm.DB, batchNumber int) error {
			for i := range terraformDeploymentBatch {
				terraformDeploymentBatch[i].LastOperationState = "failed"
				terraformDeploymentBatch[i].LastOperationMessage = "the broker restarted while the operation was in progress"
				logger.Info("mark-as-failed", lager.Data{"workspace_id": terraformDeploymentBatch[i].ID})
			}

			return tx.Save(&terraformDeploymentBatch).Error
		})

		return result.Error
	} else {
		deploymentIds, err := s.LockedDeploymentIds()
		if err != nil {
			return err
		}

		for _, id := range deploymentIds {
			var receiver models.TerraformDeployment
			if err := s.db.Where("id = ?", id).First(&receiver).Error; err != nil {
				return err
			}
			receiver.LastOperationState = "failed"
			err := s.db.Save(receiver).Error
			if err != nil {
				return err
			}
		}
		return err
	}

}
