package storage

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"gorm.io/gorm"
)

func (s *Storage) UpdateAllRecords() error {
	updaters := []func() error{
		s.updateAllSericeBindingCredentials,
		s.updateAllProvisionRequestDetails,
		s.updateAllServiceInstanceDetails,
	}
	for _, e := range updaters {
		if err := e(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) updateAllSericeBindingCredentials() error {
	var serviceBindingCredentialsBatch []models.ServiceBindingCredentials
	result := s.db.FindInBatches(&serviceBindingCredentialsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range serviceBindingCredentialsBatch {
			creds, err := s.decodeJSONObject(serviceBindingCredentialsBatch[i].OtherDetails)
			if err != nil {
				return err
			}

			serviceBindingCredentialsBatch[i].OtherDetails, err = s.encodeJSON(creds)
			if err != nil {
				return err
			}
		}

		return tx.Save(&serviceBindingCredentialsBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error re-enoding service binding credentials: %v", result.Error)
	}

	return nil
}

func (s *Storage) updateAllProvisionRequestDetails() error {
	var provisionRequestDetailsBatch []models.ProvisionRequestDetails
	result := s.db.FindInBatches(&provisionRequestDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range provisionRequestDetailsBatch {
			details, err := s.decodeBytes(provisionRequestDetailsBatch[i].RequestDetails)
			if err != nil {
				return err
			}

			provisionRequestDetailsBatch[i].RequestDetails, err = s.encodeBytes(details)
			if err != nil {
				return err
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
				return err
			}

			serviceInstanceDetailsBatch[i].OtherDetails, err = s.encodeJSON(outputs)
			if err != nil {
				return err
			}
		}

		return tx.Save(&serviceInstanceDetailsBatch).Error
	})
	if result.Error != nil {
		return fmt.Errorf("error reencrypting: %v", result.Error)
	}

	return nil
}
