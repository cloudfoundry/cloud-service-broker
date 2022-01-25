package storage

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
)

type BindRequestDetails struct {
	ServiceBindingGUID string
	RequestDetails     []byte
}

func (s *Storage) StoreBindRequestDetails(bindingID string, details json.RawMessage) error {
	if details == nil {
		return nil
	}

	encoded, err := s.encodeBytes(details)
	if err != nil {
		return fmt.Errorf("error encoding details: %w", err)
	}

	var receiver []models.BindRequestDetails
	if err := s.db.Where("service_binding_id = ?", bindingID).Find(&receiver).Error; err != nil {
		return fmt.Errorf("error searching for existing bind request details records: %w", err)
	}
	switch len(receiver) {
	case 0:
		m := models.BindRequestDetails{
			ServiceBindingId: bindingID,
			RequestDetails:   encoded,
		}
		if err := s.db.Create(&m).Error; err != nil {
			return fmt.Errorf("error creating bind request details: %w", err)
		}
	default:
		receiver[0].RequestDetails = encoded
		if err := s.db.Save(&receiver[0]).Error; err != nil {
			return fmt.Errorf("error saving bind request details: %w", err)
		}
	}

	return nil
}

func (s *Storage) GetBindRequestDetails(bindingID string) (json.RawMessage, error) {
	exists, err := s.existsBindRequestDetails(bindingID)
	switch {
	case err != nil:
		return nil, err
	case !exists:
		return nil, nil
	}

	var receiver models.BindRequestDetails
	if err := s.db.Where("service_binding_id = ?", bindingID).First(&receiver).Error; err != nil {
		return nil, fmt.Errorf("error finding bind request details record: %w", err)
	}

	decoded, err := s.decodeBytes(receiver.RequestDetails)
	if err != nil {
		return nil, fmt.Errorf("error decoding bind request details: %w", err)
	}

	return decoded, nil
}

func (s *Storage) DeleteBindRequestDetails(bindingID string) error {
	err := s.db.Where("service_binding_id = ?", bindingID).Delete(&models.BindRequestDetails{}).Error
	if err != nil {
		return fmt.Errorf("error deleting bind request details: %w", err)
	}
	return nil
}

func (s *Storage) existsBindRequestDetails(bindingID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.BindRequestDetails{}).Where("service_binding_id = ?", bindingID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("error counting bind request details: %w", err)
	}
	return count != 0, nil
}
