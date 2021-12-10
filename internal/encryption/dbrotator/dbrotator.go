package dbrotator

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"gorm.io/gorm"
)

// ReencryptDB re-encrypts the database with the primary encryptor (which can be the No-op encryptor)
func ReencryptDB(db *gorm.DB) error {
	return encryptTerraformWorkspaces(db)
}

func encryptTerraformWorkspaces(db *gorm.DB) error {
	var terraformWorkspacesBatch []models.TerraformDeployment
	result := db.FindInBatches(&terraformWorkspacesBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range terraformWorkspacesBatch {
			workspace, err := terraformWorkspacesBatch[i].GetWorkspace()
			if err != nil {
				return err
			}

			if err := terraformWorkspacesBatch[i].SetWorkspace(workspace); err != nil {
				return err
			}
		}

		return tx.Save(&terraformWorkspacesBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error reencrypting: %v", result.Error)
	}

	return nil
}
