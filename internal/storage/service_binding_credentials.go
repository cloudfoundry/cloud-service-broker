package storage

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
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
		ServiceId:         binding.ServiceGUID,
		ServiceInstanceId: binding.ServiceInstanceGUID,
		BindingId:         binding.BindingGUID,
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
		ServiceGUID:         receiver.ServiceId,
		ServiceInstanceGUID: receiver.ServiceInstanceId,
		BindingGUID:         receiver.BindingId,
		Credentials:         decoded,
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
