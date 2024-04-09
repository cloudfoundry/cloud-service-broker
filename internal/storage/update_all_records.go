package storage

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
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
			data, err := s.decodeBytes(serviceBindingCredentialsBatch[i].OtherDetails)
			if err != nil {
				return fmt.Errorf("decode error for %q: %w", serviceBindingCredentialsBatch[i].BindingID, err)
			}

			serviceBindingCredentialsBatch[i].OtherDetails, err = s.encodeBytes(data)
			if err != nil {
				return fmt.Errorf("encode error for %q: %w", serviceBindingCredentialsBatch[i].BindingID, err)
			}
		}

		return tx.Save(&serviceBindingCredentialsBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error re-encoding service binding credentials: %w", result.Error)
	}

	return nil
}

func (s *Storage) updateAllBindRequestDetails() error {
	var bindRequestDetailsBatch []models.BindRequestDetails
	result := s.db.FindInBatches(&bindRequestDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range bindRequestDetailsBatch {
			data, err := s.decodeBytes(bindRequestDetailsBatch[i].RequestDetails)
			if err != nil {
				return fmt.Errorf("decode error for %q: %w", bindRequestDetailsBatch[i].ServiceBindingID, err)
			}

			bindRequestDetailsBatch[i].RequestDetails, err = s.encodeBytes(data)
			if err != nil {
				return fmt.Errorf("encode error for %q: %w", bindRequestDetailsBatch[i].ServiceBindingID, err)
			}
		}

		return tx.Save(&bindRequestDetailsBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error re-encoding service binding request details: %w", result.Error)
	}

	return nil
}

func (s *Storage) updateAllProvisionRequestDetails() error {
	var provisionRequestDetailsBatch []models.ProvisionRequestDetails
	result := s.db.FindInBatches(&provisionRequestDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range provisionRequestDetailsBatch {
			data, err := s.decodeBytes(provisionRequestDetailsBatch[i].RequestDetails)
			if err != nil {
				return fmt.Errorf("decode error for %q: %w", provisionRequestDetailsBatch[i].ServiceInstanceID, err)
			}

			provisionRequestDetailsBatch[i].RequestDetails, err = s.encodeBytes(data)
			if err != nil {
				return fmt.Errorf("encode error for %q: %w", provisionRequestDetailsBatch[i].ServiceInstanceID, err)
			}
		}

		return tx.Save(&provisionRequestDetailsBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error re-encoding provision request details: %w", result.Error)
	}

	return nil
}

func (s *Storage) updateAllServiceInstanceDetails() error {
	var serviceInstanceDetailsBatch []models.ServiceInstanceDetails
	result := s.db.FindInBatches(&serviceInstanceDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range serviceInstanceDetailsBatch {
			data, err := s.decodeBytes(serviceInstanceDetailsBatch[i].OtherDetails)
			if err != nil {
				return fmt.Errorf("decode error for %q: %w", serviceInstanceDetailsBatch[i].ID, err)
			}

			serviceInstanceDetailsBatch[i].OtherDetails, err = s.encodeBytes(data)
			if err != nil {
				return fmt.Errorf("encode error for %q: %w", serviceInstanceDetailsBatch[i].ID, err)
			}
		}

		return tx.Save(&serviceInstanceDetailsBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error re-encoding service instance details: %w", result.Error)
	}

	return nil
}

func (s *Storage) updateAllTerraformDeployments() error {
	var terraformDeploymentBatch []models.TerraformDeployment
	result := s.db.FindInBatches(&terraformDeploymentBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range terraformDeploymentBatch {
			data, err := s.decodeBytes(terraformDeploymentBatch[i].Workspace)
			if err != nil {
				return fmt.Errorf("decode error for %q: %w", terraformDeploymentBatch[i].ID, err)
			}

			terraformDeploymentBatch[i].Workspace, err = s.encodeBytes(data)
			if err != nil {
				return fmt.Errorf("encode error for %q: %w", terraformDeploymentBatch[i].ID, err)
			}
		}

		return tx.Save(&terraformDeploymentBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error re-encoding terraform deployment: %w", result.Error)
	}

	return nil
}
