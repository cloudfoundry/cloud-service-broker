package storage

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
)

type BindRequestDetails struct {
	ServiceInstanceGUID string
	ServiceBindingGUID  string
	RequestDetails      JSONObject
}

func (s *Storage) StoreBindRequestDetails(bindingID, instanceID string, details JSONObject) error {
	// does not store if binding parameters are nil
	if details == nil {
		return nil
	}

	encoded, err := s.encodeJSON(details)
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
			ServiceInstanceID: instanceID,
			ServiceBindingID:  bindingID,
			RequestDetails:    encoded,
		}
		if err := s.db.Create(&m).Error; err != nil {
			return fmt.Errorf("error creating bind request details: %w", err)
		}
	default:
		return fmt.Errorf("error saving bind request details: Binding already exists: %w", err)
	}

	return nil
}

func (s *Storage) GetBindRequestDetails(bindingID string, instanceID string) (JSONObject, error) {
	exists, err := s.existsBindRequestDetails(bindingID, instanceID)
	switch {
	case err != nil:
		return nil, err
	case !exists:
		return nil, nil
	}

	var receiver models.BindRequestDetails
	if err := s.db.Where("service_binding_id = ? AND service_instance_id = ?", bindingID, instanceID).First(&receiver).Error; err != nil {
		return nil, fmt.Errorf("error finding bind request details record: %w", err)
	}

	decoded, err := s.decodeJSONObject(receiver.RequestDetails)
	if err != nil {
		return nil, fmt.Errorf("error decoding bind request details %q: %w", bindingID, err)
	}

	return decoded, nil
}

func (s *Storage) DeleteBindRequestDetails(bindingID string, instanceID string) error {
	err := s.db.Where("service_binding_id = ? AND service_instance_id = ?", bindingID, instanceID).Unscoped().Delete(&models.BindRequestDetails{}).Error
	if err != nil {
		return fmt.Errorf("error deleting bind request details: %w", err)
	}
	return nil
}

func (s *Storage) existsBindRequestDetails(bindingID, instanceID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.BindRequestDetails{}).Where("service_binding_id = ? AND service_instance_id = ?", bindingID, instanceID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("error counting bind request details: %w", err)
	}
	return count != 0, nil
}
