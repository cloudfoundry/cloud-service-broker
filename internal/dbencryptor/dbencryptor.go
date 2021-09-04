package dbencryptor

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"gorm.io/gorm"
)

// EncryptDB encrypts the database with the primary encryptor (which can be the No-op encryptor)
func EncryptDB(db *gorm.DB) error {
	dbEncryptors := []func(*gorm.DB) error{
		encryptProvisionRequestDetails,
		encryptServiceInstanceDetails,
		encryptServiceBindingCredentials,
		encryptTerraformWorkspaces,
	}
	for _, e := range dbEncryptors {
		if err := e(db); err != nil {
			return nil
		}
	}

	return nil
}

func encryptProvisionRequestDetails(db *gorm.DB) error {
	var provisionRequestDetailsBatch []models.ProvisionRequestDetails
	db.FindInBatches(&provisionRequestDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range provisionRequestDetailsBatch {
			details, err := provisionRequestDetailsBatch[i].GetRequestDetails()
			if err != nil {
				return err
			}

			if err := provisionRequestDetailsBatch[i].SetRequestDetails(details); err != nil {
				return err
			}
		}

		return tx.Save(&provisionRequestDetailsBatch).Error
	})

	return nil
}

func encryptServiceInstanceDetails(db *gorm.DB) error {
	var serviceInstanceDetailsBatch []models.ServiceInstanceDetails
	db.FindInBatches(&serviceInstanceDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range serviceInstanceDetailsBatch {
			var details interface{}
			if err := serviceInstanceDetailsBatch[i].GetOtherDetails(&details); err != nil {
				return err
			}

			if err := serviceInstanceDetailsBatch[i].SetOtherDetails(details); err != nil {
				return err
			}
		}

		return tx.Save(&serviceInstanceDetailsBatch).Error
	})

	return nil
}

func encryptServiceBindingCredentials(db *gorm.DB) error {
	var serviceBindingCredentialsBatch []models.ServiceBindingCredentials
	db.FindInBatches(&serviceBindingCredentialsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range serviceBindingCredentialsBatch {
			var details interface{}
			if err := serviceBindingCredentialsBatch[i].GetOtherDetails(&details); err != nil {
				return err
			}

			if err := serviceBindingCredentialsBatch[i].SetOtherDetails(details); err != nil {
				return err
			}
		}

		return tx.Save(&serviceBindingCredentialsBatch).Error
	})

	return nil
}

func encryptTerraformWorkspaces(db *gorm.DB) error {
	var terraformWorkspacesBatch []models.TerraformDeployment
	db.FindInBatches(&terraformWorkspacesBatch, 100, func(tx *gorm.DB, batchNumber int) error {
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

	return nil
}
