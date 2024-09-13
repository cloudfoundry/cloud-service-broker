package storage

import (
	"os"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"gorm.io/gorm"
)

const FailedMessage = "the broker restarted while the operation was in progress"

func (s *Storage) RecoverInProgressOperations(logger lager.Logger) error {
	logger = logger.Session("recover-in-progress-operations")

	if runningAsCFApp() {
		return s.markAllInProgressOperationsAsFailed(logger)
	} else {
		return s.markAllOperationsWithLockFilesAsFailed(logger)
	}

}

func runningAsCFApp() bool {
	return os.Getenv("CF_INSTANCE_GUID") != ""
}

func (s *Storage) markAllInProgressOperationsAsFailed(logger lager.Logger) error {
	logger.Info("checking all in in progress operations from DB")
	var terraformDeploymentBatch []models.TerraformDeployment
	result := s.db.Where("last_operation_state = ?", "in progress").FindInBatches(&terraformDeploymentBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range terraformDeploymentBatch {
			terraformDeploymentBatch[i].LastOperationState = "failed"
			terraformDeploymentBatch[i].LastOperationMessage = FailedMessage
			logger.Info("mark-as-failed", lager.Data{"workspace_id": terraformDeploymentBatch[i].ID})
		}

		return tx.Save(&terraformDeploymentBatch).Error
	})

	return result.Error
}

func (s *Storage) markAllOperationsWithLockFilesAsFailed(logger lager.Logger) error {
	logger.Info("checking all in in progress operations from lockfiles")

	deploymentIds, err := s.GetLockedDeploymentIds()
	if err != nil {
		return err
	}

	for _, id := range deploymentIds {
		var receiver models.TerraformDeployment
		if err := s.db.Where("id = ?", id).First(&receiver).Error; err != nil {
			return err
		}
		receiver.LastOperationState = "failed"
		receiver.LastOperationMessage = FailedMessage

		err := s.db.Save(receiver).Error
		if err != nil {
			return err
		}
		logger.Info("mark-as-failed", lager.Data{"workspace_id": id})
	}
	return err
}
