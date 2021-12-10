package storage

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"gorm.io/gorm"
)

type Credentials map[string]interface{}

type ServiceBindingCredentials struct {
	ServiceID         string
	ServiceInstanceID string
	BindingID         string
	Credentials       Credentials
}

func (s *Storage) CreateServiceBindingCredentials(binding ServiceBindingCredentials) error {
	encodedCreds, err := s.marshalAndEncrypt(binding.Credentials)
	if err != nil {
		return fmt.Errorf("error encoding credentials: %w", err)
	}

	m := models.ServiceBindingCredentials{
		OtherDetails:      encodedCreds,
		ServiceId:         binding.ServiceID,
		ServiceInstanceId: binding.ServiceInstanceID,
		BindingId:         binding.BindingID,
	}

	if err := s.db.Create(&m).Error; err != nil {
		return fmt.Errorf("error creating service credential binding: %w", err)
	}

	return nil
}

func (s *Storage) GetServiceBindingCredentials(bindingID, serviceInstanceID string) (ServiceBindingCredentials, error) {
	var receiver models.ServiceBindingCredentials
	if err := s.db.Where("service_instance_id = ? AND binding_id = ?", serviceInstanceID, bindingID).First(&receiver).Error; err != nil {
		return ServiceBindingCredentials{}, fmt.Errorf("error finding service credential binding: %w", err)
	}

	decoded, err := s.decryptAndUnmarshalObject(receiver.OtherDetails)
	if err != nil {
		return ServiceBindingCredentials{}, fmt.Errorf("error decoding credentials: %w", err)
	}

	return ServiceBindingCredentials{
		ServiceID:         receiver.ServiceId,
		ServiceInstanceID: receiver.ServiceInstanceId,
		BindingID:         receiver.BindingId,
		Credentials:       decoded,
	}, nil
}

func (s *Storage) ExistsServiceBindingCredentials(bindingID, serviceInstanceID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.ServiceBindingCredentials{}).Where("service_instance_id = ? AND binding_id = ?", serviceInstanceID, bindingID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("error counting service credential binding: %w", err)
	}
	return count != 0, nil
}

func (s *Storage) DeleteServiceBindingCredentials(bindingID, serviceInstanceID string) error {
	err := s.db.Where("service_instance_id = ? AND binding_id = ?", serviceInstanceID, bindingID).Delete(&models.ServiceBindingCredentials{}).Error
	if err != nil {
		return fmt.Errorf("error deleting service binding credentials: %w", err)
	}
	return nil
}

func (s *Storage) UpdateAllServiceBindingCredentials() error {
	var serviceBindingCredentialsBatch []models.ServiceBindingCredentials
	result := s.db.FindInBatches(&serviceBindingCredentialsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range serviceBindingCredentialsBatch {
			creds, err := s.decryptAndUnmarshalObject(serviceBindingCredentialsBatch[i].OtherDetails)
			if err != nil {
				return err
			}

			serviceBindingCredentialsBatch[i].OtherDetails, err = s.marshalAndEncrypt(creds)
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
