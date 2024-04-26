package storage

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/v3/dbservice/models"
)

type ServiceBindingCredentials struct {
	ServiceGUID         string
	ServiceInstanceGUID string
	BindingGUID         string
	Credentials         JSONObject
}

func (s *Storage) CreateServiceBindingCredentials(binding ServiceBindingCredentials) error {
	encodedCreds, err := s.encodeJSON(binding.Credentials)
	if err != nil {
		return fmt.Errorf("error encoding credentials: %w", err)
	}

	m := models.ServiceBindingCredentials{
		OtherDetails:      encodedCreds,
		ServiceID:         binding.ServiceGUID,
		ServiceInstanceID: binding.ServiceInstanceGUID,
		BindingID:         binding.BindingGUID,
	}

	if err := s.db.Create(&m).Error; err != nil {
		return fmt.Errorf("error creating service credential binding: %w", err)
	}

	return nil
}

func (s *Storage) GetServiceBindingCredentials(bindingID, serviceInstanceID string) (ServiceBindingCredentials, error) {
	exists, err := s.ExistsServiceBindingCredentials(bindingID, serviceInstanceID)
	switch {
	case err != nil:
		return ServiceBindingCredentials{}, err
	case !exists:
		return ServiceBindingCredentials{}, fmt.Errorf("could not find binding credentials for binding %q and service instance %q", bindingID, serviceInstanceID)
	}

	var receiver models.ServiceBindingCredentials
	if err := s.db.Where("service_instance_id = ? AND binding_id = ?", serviceInstanceID, bindingID).First(&receiver).Error; err != nil {
		return ServiceBindingCredentials{}, fmt.Errorf("error finding service credential binding: %w", err)
	}

	decoded, err := s.decodeJSONObject(receiver.OtherDetails)
	if err != nil {
		return ServiceBindingCredentials{}, fmt.Errorf("error decoding binding credentials %q: %w", bindingID, err)
	}

	return ServiceBindingCredentials{
		ServiceGUID:         receiver.ServiceID,
		ServiceInstanceGUID: receiver.ServiceInstanceID,
		BindingGUID:         receiver.BindingID,
		Credentials:         decoded,
	}, nil
}

func (s *Storage) GetServiceBindingIDsForServiceInstance(serviceInstanceID string) ([]string, error) {
	var receiver []models.ServiceBindingCredentials
	if err := s.db.Where("service_instance_id = ?", serviceInstanceID).Find(&receiver).Error; err != nil {
		return nil, err
	}

	result := make([]string, len(receiver))
	for i := range receiver {
		result[i] = receiver[i].BindingID
	}

	return result, nil
}

func (s *Storage) ExistsServiceBindingCredentials(bindingID, serviceInstanceID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.ServiceBindingCredentials{}).Where("service_instance_id = ? AND binding_id = ?", serviceInstanceID, bindingID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("error counting service credential binding: %w", err)
	}
	return count != 0, nil
}

func (s *Storage) DeleteServiceBindingCredentials(bindingID, serviceInstanceID string) error {
	err := s.db.Where("service_instance_id = ? AND binding_id = ?", serviceInstanceID, bindingID).Unscoped().Delete(&models.ServiceBindingCredentials{}).Error
	if err != nil {
		return fmt.Errorf("error deleting service binding credentials: %w", err)
	}
	return nil
}
