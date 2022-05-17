package storage

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
)

type TerraformDeployment struct {
	ID                   string
	Workspace            workspace.Workspace
	LastOperationType    string
	LastOperationState   string
	LastOperationMessage string
}

func (deployment *TerraformDeployment) TFWorkspace() *workspace.TerraformWorkspace {
	return deployment.Workspace.(*workspace.TerraformWorkspace)
}

func (s *Storage) StoreTerraformDeployment(t TerraformDeployment) error {
	encoded, err := s.encodeJSON(t.Workspace)
	if err != nil {
		return fmt.Errorf("error encoding workspace: %w", err)
	}

	var m models.TerraformDeployment
	if err := s.loadTerraformDeploymentIfExists(t.ID, &m); err != nil {
		return err
	}

	m.Workspace = encoded
	m.LastOperationType = t.LastOperationType
	m.LastOperationState = t.LastOperationState
	m.LastOperationMessage = t.LastOperationMessage

	switch m.ID {
	case "":
		m.ID = t.ID
		if err := s.db.Create(&m).Error; err != nil {
			return fmt.Errorf("error creating terraform deployment: %w", err)
		}
	default:
		if err := s.db.Save(&m).Error; err != nil {
			return fmt.Errorf("error saving terraform deployment: %w", err)
		}
	}

	return nil
}

func (s *Storage) GetTerraformDeployment(id string) (TerraformDeployment, error) {
	exists, err := s.ExistsTerraformDeployment(id)
	switch {
	case err != nil:
		return TerraformDeployment{}, err
	case !exists:
		return TerraformDeployment{}, fmt.Errorf("could not find terraform deployment: %s", id)
	}

	var receiver models.TerraformDeployment
	if err := s.db.Where("id = ?", id).First(&receiver).Error; err != nil {
		return TerraformDeployment{}, fmt.Errorf("error finding terraform deployment: %w", err)
	}

	var tfWorkspace workspace.TerraformWorkspace
	if err = s.decodeJSON(receiver.Workspace, &tfWorkspace); err != nil {
		return TerraformDeployment{}, fmt.Errorf("error decoding workspace %q: %w", id, err)
	}
	return TerraformDeployment{
		ID:                   id,
		LastOperationType:    receiver.LastOperationType,
		LastOperationState:   receiver.LastOperationState,
		LastOperationMessage: receiver.LastOperationMessage,
		Workspace:            &tfWorkspace,
	}, nil
}

func (s *Storage) ExistsTerraformDeployment(id string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.TerraformDeployment{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("error counting terraform deployments: %w", err)
	}
	return count != 0, nil
}

func (s *Storage) DeleteTerraformDeployment(id string) error {
	err := s.db.Where("id = ?", id).Delete(&models.TerraformDeployment{}).Error
	if err != nil {
		return fmt.Errorf("error deleting terraform deployment: %w", err)
	}
	return nil
}

func (s *Storage) loadTerraformDeploymentIfExists(id string, receiver interface{}) error {
	if id == "" {
		return nil
	}

	exists, err := s.ExistsTerraformDeployment(id)
	switch {
	case err != nil:
		return err
	case !exists:
		return nil
	}

	return s.db.Where("id = ?", id).First(receiver).Error
}
