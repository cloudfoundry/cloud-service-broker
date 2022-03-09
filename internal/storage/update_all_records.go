package storage

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	"gorm.io/gorm"
)

func (s *Storage) UpdateAllRecords() error {
	updaters := []func() error{
		s.updateAllServiceBindingCredentials,
		s.updateAllBindRequestDetails,
		s.updateAllProvisionRequestDetails,
		s.updateAllServiceInstanceDetails,
		s.updateAllTerraformDeployments,
	}
	for _, e := range updaters {
		if err := e(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) updateAllServiceBindingCredentials() error {
	var serviceBindingCredentialsBatch []models.ServiceBindingCredentials
	result := s.db.FindInBatches(&serviceBindingCredentialsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range serviceBindingCredentialsBatch {
			creds, err := s.decodeJSONObject(serviceBindingCredentialsBatch[i].OtherDetails)
			if err != nil {
				return fmt.Errorf("decode error for %q: %w", serviceBindingCredentialsBatch[i].BindingId, err)
			}

			serviceBindingCredentialsBatch[i].OtherDetails, err = s.encodeJSON(creds)
			if err != nil {
				return fmt.Errorf("encode error for %q: %w", serviceBindingCredentialsBatch[i].BindingId, err)
			}
		}

		return tx.Save(&serviceBindingCredentialsBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error re-encoding service binding credentials: %v", result.Error)
	}

	return nil
}

func (s *Storage) updateAllBindRequestDetails() error {
	var bindRequestDetailsBatch []models.BindRequestDetails
	result := s.db.FindInBatches(&bindRequestDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range bindRequestDetailsBatch {
			details, err := s.decodeBytes(bindRequestDetailsBatch[i].RequestDetails)
			if err != nil {
				return fmt.Errorf("decode error for %q: %w", bindRequestDetailsBatch[i].ServiceBindingId, err)
			}

			bindRequestDetailsBatch[i].RequestDetails, err = s.encodeBytes(details)
			if err != nil {
				return fmt.Errorf("encode error for %q: %w", bindRequestDetailsBatch[i].ServiceBindingId, err)
			}
		}

		return tx.Save(&bindRequestDetailsBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error re-encoding service binding request details: %v", result.Error)
	}

	return nil
}

func (s *Storage) updateAllProvisionRequestDetails() error {
	var provisionRequestDetailsBatch []models.ProvisionRequestDetails
	result := s.db.FindInBatches(&provisionRequestDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range provisionRequestDetailsBatch {
			details, err := s.decodeBytes(provisionRequestDetailsBatch[i].RequestDetails)
			if err != nil {
				return fmt.Errorf("decode error for %q: %w", provisionRequestDetailsBatch[i].ServiceInstanceId, err)
			}

			provisionRequestDetailsBatch[i].RequestDetails, err = s.encodeBytes(details)
			if err != nil {
				return fmt.Errorf("encode error for %q: %w", provisionRequestDetailsBatch[i].ServiceInstanceId, err)
			}
		}

		return tx.Save(&provisionRequestDetailsBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error re-encoding provision request details: %v", result.Error)
	}

	return nil
}

func (s *Storage) updateAllServiceInstanceDetails() error {
	var serviceInstanceDetailsBatch []models.ServiceInstanceDetails
	result := s.db.FindInBatches(&serviceInstanceDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range serviceInstanceDetailsBatch {
			outputs, err := s.decodeJSONObject(serviceInstanceDetailsBatch[i].OtherDetails)
			if err != nil {
				return fmt.Errorf("decode error for %q: %w", serviceInstanceDetailsBatch[i].ID, err)
			}

			serviceInstanceDetailsBatch[i].OtherDetails, err = s.encodeJSON(outputs)
			if err != nil {
				return fmt.Errorf("encode error for %q: %w", serviceInstanceDetailsBatch[i].ID, err)
			}
		}

		return tx.Save(&serviceInstanceDetailsBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error re-encoding service instance details: %v", result.Error)
	}

	return nil
}

func (s *Storage) updateAllTerraformDeployments() error {
	var terraformDeploymentBatch []models.TerraformDeployment
	result := s.db.FindInBatches(&terraformDeploymentBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range terraformDeploymentBatch {
			workspace, err := s.decodeBytes(terraformDeploymentBatch[i].Workspace)
			if err != nil {
				return fmt.Errorf("decode error for %q: %w", terraformDeploymentBatch[i].ID, err)
			}

			terraformDeploymentBatch[i].Workspace, err = s.encodeBytes(workspace)
			if err != nil {
				return fmt.Errorf("encode error for %q: %w", terraformDeploymentBatch[i].ID, err)
			}
		}

		return tx.Save(&terraformDeploymentBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error re-encoding terraform deployment: %v", result.Error)
	}

	return nil
}
