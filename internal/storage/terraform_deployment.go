package storage

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"gorm.io/gorm"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace"
)

type TerraformDeployment struct {
	ID                   string
	Workspace            workspace.Workspace
	LastOperationType    string
	LastOperationState   string
	LastOperationMessage string
}

type TerraformDeploymentListEntry struct {
	ID                   string
	LastOperationType    string
	LastOperationState   string
	LastOperationMessage string
	StateVersion         *version.Version
	UpdatedAt            time.Time
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

func (s *Storage) GetAllTerraformDeployments() ([]TerraformDeploymentListEntry, error) {
	var result []TerraformDeploymentListEntry

	var terraformDeploymentBatch []models.TerraformDeployment
	status := s.db.FindInBatches(&terraformDeploymentBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range terraformDeploymentBatch {
			var tfWorkspace workspace.TerraformWorkspace
			if err := s.decodeJSON(terraformDeploymentBatch[i].Workspace, &tfWorkspace); err != nil {
				return fmt.Errorf("error decoding workspace %q: %w", terraformDeploymentBatch[i].ID, err)
			}

			tfVersion, err := tfWorkspace.StateTFVersion()
			if err != nil {
				tfVersion = nil
			}

			result = append(result, TerraformDeploymentListEntry{
				ID:                   terraformDeploymentBatch[i].ID,
				LastOperationType:    terraformDeploymentBatch[i].LastOperationType,
				LastOperationState:   terraformDeploymentBatch[i].LastOperationState,
				LastOperationMessage: terraformDeploymentBatch[i].LastOperationMessage,
				StateVersion:         tfVersion,
				UpdatedAt:            terraformDeploymentBatch[i].UpdatedAt,
			})
		}

		return nil
	})
	if status.Error != nil {
		return nil, fmt.Errorf("error reading terraform deployment batch: %w", status.Error)
	}

	return result, nil
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

func (s *Storage) loadTerraformDeploymentIfExists(id string, receiver any) error {
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

func (s *Storage) LockFilesExist() bool {
	entries, _ := os.ReadDir(s.lockFileDir)
	return len(entries) != 0
}

func (s *Storage) WriteLockFile(deploymentID string) error {
	return os.WriteFile(fmt.Sprintf("%s/%s", s.lockFileDir, sanitizeFileName(deploymentID)), []byte{}, 0o644)
}

func (s *Storage) RemoveLockFile(deploymentID string) error {
	return os.Remove(fmt.Sprintf("%s/%s", s.lockFileDir, sanitizeFileName(deploymentID)))
}

func (s *Storage) GetLockedDeploymentIds() ([]string, error) {
	entries, err := os.ReadDir(s.lockFileDir)
	var names []string
	for _, entry := range entries {
		names = append(names, strings.ReplaceAll(entry.Name(), "_", ":"))
	}
	return names, err
}

func sanitizeFileName(name string) string {
	return strings.ReplaceAll(name, ":", "_")
}
