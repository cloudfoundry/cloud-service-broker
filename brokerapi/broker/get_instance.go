package broker

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager/v3"
	"github.com/pivotal-cf/brokerapi/v11/domain"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/correlation"
)

// GetInstance fetches information about a service instance
// GET /v2/service_instances/{instance_id}
func (broker *ServiceBroker) GetInstance(ctx context.Context, instanceID string, details domain.FetchInstanceDetails) (domain.GetInstanceDetailsSpec, error) {
	broker.Logger.Info("GetInstance", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
		"service_id":  details.ServiceID,
		"plan_id":     details.PlanID,
	})

	// check whether instance exists
	exists, err := broker.store.ExistsServiceInstanceDetails(instanceID)
	if err != nil {
		return domain.GetInstanceDetailsSpec{}, fmt.Errorf("error checking for existing instance: %w", err)
	}
	if !exists {
		return domain.GetInstanceDetailsSpec{}, ErrNotFound
	}

	// get instance details
	instanceRecord, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return domain.GetInstanceDetailsSpec{}, fmt.Errorf("error retrieving service instance details: %w", err)
	}

	// check whether request parameters (if not empty) match instance details
	if len(details.ServiceID) > 0 && details.ServiceID != instanceRecord.ServiceGUID {
		return domain.GetInstanceDetailsSpec{}, ErrNotFound
	}
	if len(details.PlanID) > 0 && details.PlanID != instanceRecord.PlanGUID {
		return domain.GetInstanceDetailsSpec{}, ErrNotFound
	}

	// get instance status
	_, serviceProvider, err := broker.getDefinitionAndProvider(instanceRecord.ServiceGUID)
	if err != nil {
		return domain.GetInstanceDetailsSpec{}, fmt.Errorf("error retrieving service definition: %w", err)
	}

	done, _, lastOperationType, err := serviceProvider.PollInstance(ctx, instanceRecord.GUID)
	if err != nil {
		return domain.GetInstanceDetailsSpec{}, fmt.Errorf("error polling instance status: %w", err)
	}

	switch lastOperationType {
	case models.ProvisionOperationType:
		if !done {
			return domain.GetInstanceDetailsSpec{}, ErrNotFound
		}
	case models.UpdateOperationType:
		if !done {
			return domain.GetInstanceDetailsSpec{}, ErrConcurrencyError
		}
	}

	// get provision parameters
	params, err := broker.store.GetProvisionRequestDetails(instanceID)
	if err != nil {
		return domain.GetInstanceDetailsSpec{}, fmt.Errorf("error retrieving provision request details: %w", err)
	}

	return domain.GetInstanceDetailsSpec{
		ServiceID:    instanceRecord.ServiceGUID,
		PlanID:       instanceRecord.PlanGUID,
		DashboardURL: "",
		Parameters:   params,
		Metadata:     domain.InstanceMetadata{},
	}, nil
}
