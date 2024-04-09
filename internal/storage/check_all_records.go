package storage

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"gorm.io/gorm"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace"
)

func (s *Storage) CheckAllRecords() error {
	var errs *multierror.Error

	checkers := []func() *multierror.Error{
		s.checkAllServiceBindingCredentials,
		s.checkAllBindRequestDetails,
		s.checkAllProvisionRequestDetails,
		s.checkAllServiceInstanceDetails,
		s.checkAllTerraformDeployments,
	}
	for _, e := range checkers {
		if err := e(); err != nil {
			errs = multierror.Append(err, errs)
		}
	}

	return errs.ErrorOrNil()
}

func (s *Storage) checkAllServiceBindingCredentials() (errs *multierror.Error) {
	var serviceBindingCredentialsBatch []models.ServiceBindingCredentials
	result := s.db.FindInBatches(&serviceBindingCredentialsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range serviceBindingCredentialsBatch {
			if _, err := s.decodeJSONObject(serviceBindingCredentialsBatch[i].OtherDetails); err != nil {
				errs = multierror.Append(fmt.Errorf("decode error for service binding credential %q: %w", serviceBindingCredentialsBatch[i].BindingID, err), errs)
			}
		}

		return nil
	})
	if result.Error != nil {
		errs = multierror.Append(fmt.Errorf("error reading service binding credentials: %w", result.Error), errs)
	}

	return errs
}

func (s *Storage) checkAllBindRequestDetails() (errs *multierror.Error) {
	var bindRequestDetailsBatch []models.BindRequestDetails
	result := s.db.FindInBatches(&bindRequestDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range bindRequestDetailsBatch {
			if _, err := s.decodeJSONObject(bindRequestDetailsBatch[i].RequestDetails); err != nil {
				errs = multierror.Append(fmt.Errorf("decode error for binding request details %q: %w", bindRequestDetailsBatch[i].ServiceBindingID, err), errs)
			}
		}

		return nil
	})
	if result.Error != nil {
		errs = multierror.Append(fmt.Errorf("error reading binding request details: %w", result.Error), errs)
	}

	return errs
}

func (s *Storage) checkAllProvisionRequestDetails() (errs *multierror.Error) {
	var provisionRequestDetailsBatch []models.ProvisionRequestDetails
	result := s.db.FindInBatches(&provisionRequestDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range provisionRequestDetailsBatch {
			if _, err := s.decodeJSONObject(provisionRequestDetailsBatch[i].RequestDetails); err != nil {
				errs = multierror.Append(fmt.Errorf("decode error for provision request details %q: %w", provisionRequestDetailsBatch[i].ServiceInstanceID, err), errs)
			}
		}

		return nil
	})
	if result.Error != nil {
		errs = multierror.Append(fmt.Errorf("error reading provision request details: %w", result.Error), errs)
	}

	return errs
}

func (s *Storage) checkAllServiceInstanceDetails() (errs *multierror.Error) {
	var serviceInstanceDetailsBatch []models.ServiceInstanceDetails
	result := s.db.FindInBatches(&serviceInstanceDetailsBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range serviceInstanceDetailsBatch {
			if _, err := s.decodeJSONObject(serviceInstanceDetailsBatch[i].OtherDetails); err != nil {
				errs = multierror.Append(fmt.Errorf("decode error for service instance details %q: %w", serviceInstanceDetailsBatch[i].ID, err), errs)
			}
		}

		return nil
	})
	if result.Error != nil {
		errs = multierror.Append(fmt.Errorf("error re-encoding service instance details: %w", result.Error), errs)
	}

	return errs
}

func (s *Storage) checkAllTerraformDeployments() (errs *multierror.Error) {
	var terraformDeploymentBatch []models.TerraformDeployment
	result := s.db.FindInBatches(&terraformDeploymentBatch, 100, func(tx *gorm.DB, batchNumber int) error {
		for i := range terraformDeploymentBatch {
			var tfWorkspace workspace.TerraformWorkspace
			if err := s.decodeJSON(terraformDeploymentBatch[i].Workspace, &tfWorkspace); err != nil {
				errs = multierror.Append(fmt.Errorf("decode error for terraform deployment %q: %w", terraformDeploymentBatch[i].ID, err), errs)
			}
		}

		return nil
	})
	if result.Error != nil {
		errs = multierror.Append(fmt.Errorf("error re-encoding terraform deployment: %w", result.Error), errs)
	}

	return errs
}
