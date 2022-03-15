package storage

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
)

func (s *Storage) StoreProvisionRequestDetails(serviceInstanceID string, details JSONObject) error {
	encoded, err := s.encodeJSON(details)
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

func (s *Storage) GetProvisionRequestDetails(serviceInstanceID string) (JSONObject, error) {
	exists, err := s.existsProvisionRequestDetails(serviceInstanceID)
	switch {
	case err != nil:
		return nil, err
	case !exists:
		return nil, fmt.Errorf("could not find provision request details for service instance: %s", serviceInstanceID)
	}

	var receiver models.ProvisionRequestDetails
	if err := s.db.Where("service_instance_id = ?", serviceInstanceID).First(&receiver).Error; err != nil {
		return nil, fmt.Errorf("error finding provision request details record: %w", err)
	}

	decoded, err := s.decodeJSONObject(receiver.RequestDetails)
	if err != nil {
		return nil, fmt.Errorf("error decoding provision request details %q: %w", serviceInstanceID, err)
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

func (s *Storage) existsProvisionRequestDetails(serviceInstanceID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.ProvisionRequestDetails{}).Where("service_instance_id = ?", serviceInstanceID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("error counting provision request details: %w", err)
	}
	return count != 0, nil
}
