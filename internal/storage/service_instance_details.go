package storage

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"gorm.io/gorm"
)

type ServiceInstanceDetails struct {
	GUID             string
	Name             string
	Location         string
	URL              string
	Outputs          JSONObject
	ServiceGUID      string
	PlanGUID         string
	SpaceGUID        string
	OrganizationGUID string
	OperationType    string
	OperationGUID    string
}

func (s *Storage) StoreServiceInstanceDetails(d ServiceInstanceDetails) error {
	encoded, err := s.encodeJSON(d.Outputs)
	if err != nil {
		return fmt.Errorf("error encoding details: %w", err)
	}

	var m models.ServiceInstanceDetails
	if err := s.loadServiceInstanceDetailsIfExists(d.GUID, &m); err != nil {
		return err
	}

	m.Name = d.Name
	m.Location = d.Location
	m.URL = d.URL
	m.OtherDetails = encoded
	m.ServiceID = d.ServiceGUID
	m.PlanID = d.PlanGUID
	m.SpaceGUID = d.SpaceGUID
	m.OrganizationGUID = d.OrganizationGUID
	m.OperationType = d.OperationType
	m.OperationID = d.OperationGUID

	switch m.ID {
	case "":
		m.ID = d.GUID
		if err := s.db.Create(&m).Error; err != nil {
			return fmt.Errorf("error creating service instance details: %w", err)
		}
	default:
		if err := s.db.Save(&m).Error; err != nil {
			return fmt.Errorf("error saving service instance details: %w", err)
		}
	}

	return nil
}

func (s *Storage) ExistsServiceInstanceDetails(guid string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.ServiceInstanceDetails{}).Where("id = ?", guid).Count(&count).Error; err != nil {
		return false, fmt.Errorf("error counting service instance details: %w", err)
	}
	return count != 0, nil
}

func (s *Storage) GetServiceInstanceDetails(guid string) (ServiceInstanceDetails, error) {
	exists, err := s.ExistsServiceInstanceDetails(guid)
	switch {
	case err != nil:
		return ServiceInstanceDetails{}, err
	case !exists:
		return ServiceInstanceDetails{}, fmt.Errorf("could not find service instance details for: %s", guid)
	}

	var receiver models.ServiceInstanceDetails
	if err := s.db.Where("id = ?", guid).First(&receiver).Error; err != nil {
		return ServiceInstanceDetails{}, fmt.Errorf("error finding service instance details: %w", err)
	}

	decoded, err := s.decodeJSONObject(receiver.OtherDetails)
	if err != nil {
		return ServiceInstanceDetails{}, fmt.Errorf("error decoding service instance outputs %q: %w", guid, err)
	}

	return ServiceInstanceDetails{
		GUID:             guid,
		Name:             receiver.Name,
		Location:         receiver.Location,
		URL:              receiver.URL,
		Outputs:          decoded,
		ServiceGUID:      receiver.ServiceID,
		PlanGUID:         receiver.PlanID,
		SpaceGUID:        receiver.SpaceGUID,
		OrganizationGUID: receiver.OrganizationGUID,
		OperationType:    receiver.OperationType,
		OperationGUID:    receiver.OperationID,
	}, nil
}

func (s *Storage) GetServiceInstancesIDs() (ids []string, err error) {
	var serviceInstanceDetailsBatch []models.ServiceInstanceDetails
	result := s.db.FindInBatches(&serviceInstanceDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for _, e := range serviceInstanceDetailsBatch {
			ids = append(ids, e.ID)
		}

		return nil
	})
	if result.Error != nil {
		return nil, fmt.Errorf("error reading service instance ids: %w", result.Error)
	}

	return ids, nil
}

func (s *Storage) DeleteServiceInstanceDetails(guid string) error {
	err := s.db.Where("id = ?", guid).Delete(&models.ServiceInstanceDetails{}).Error
	if err != nil {
		return fmt.Errorf("error deleting service instance details: %w", err)
	}
	return nil
}

func (s *Storage) loadServiceInstanceDetailsIfExists(guid string, receiver any) error {
	if guid == "" {
		return nil
	}

	exists, err := s.ExistsServiceInstanceDetails(guid)
	switch {
	case err != nil:
		return err
	case !exists:
		return nil
	}

	return s.db.Where("id = ?", guid).First(receiver).Error
}
