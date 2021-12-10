package storage

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
)

type ProvisionRequestDetails struct {
	ServiceInstanceId string
	RequestDetails    []byte
}

func (s *Storage) StoreProvisionRequestDetails(serviceInstanceID string, details json.RawMessage) error {
	encoded, err := s.encodeBytes(details)
	if err != nil {
		return fmt.Errorf("error encoding details: %w", err)
	}

	var receiver []models.ProvisionRequestDetails
	if err := s.db.Where("service_instance_id = ?", serviceInstanceID).Find(&receiver).Error; err != nil {
		return fmt.Errorf("error searching for existing provision request details records: %w", err)
	}
	switch len(receiver) {
	case 0:
		m := models.ProvisionRequestDetails{
			ServiceInstanceId: serviceInstanceID,
			RequestDetails:    encoded,
		}
		if err := s.db.Create(&m).Error; err != nil {
			return fmt.Errorf("error creating provision request details: %w", err)
		}
	default:
		receiver[0].RequestDetails = encoded
		if err := s.db.Save(&receiver[0]).Error; err != nil {
			return fmt.Errorf("error saving provision request details: %w", err)
		}
	}

	return nil
}

func (s *Storage) GetProvisionRequestDetails(serviceInstanceID string) (json.RawMessage, error) {
	var receiver models.ProvisionRequestDetails
	if err := s.db.Where("service_instance_id = ?", serviceInstanceID).First(&receiver).Error; err != nil {
		return nil, fmt.Errorf("error finding provision request details record: %w", err)
	}

	decoded, err := s.decodeBytes(receiver.RequestDetails)
	if err != nil {
		return nil, fmt.Errorf("error decoding provision request details: %w", err)
	}

	return decoded, nil
}

func (s *Storage) DeleteProvisionRequestDetails(serviceInstanceID string) error {
	err := s.db.Where("service_instance_id = ?", serviceInstanceID).Delete(&models.ProvisionRequestDetails{}).Error
	if err != nil {
		return fmt.Errorf("error deleting provision request details: %w", err)
	}
	return nil
}
