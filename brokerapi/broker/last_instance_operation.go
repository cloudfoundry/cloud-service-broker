package broker

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

// LastOperation fetches last operation state for a service instance.
// It is bound to the `GET /v2/service_instances/:instance_id/last_operation` endpoint.
// It is called by `cf create-service` or `cf delete-service` if the operation was asynchronous.
func (broker *ServiceBroker) LastOperation(ctx context.Context, instanceID string, details domain.PollDetails) (domain.LastOperation, error) {
	broker.Logger.Info("Last Operation", correlation.ID(ctx), lager.Data{
		"instance_id":    instanceID,
		"plan_id":        details.PlanID,
		"service_id":     details.ServiceID,
		"operation_data": details.OperationData,
	})

	instance, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return domain.LastOperation{}, apiresponses.ErrInstanceDoesNotExist
	}

	_, serviceProvider, err := broker.getDefinitionAndProvider(instance.ServiceGUID)
	if err != nil {
		return domain.LastOperation{}, err
	}

	lastOperationType := instance.OperationType

	done, message, err := serviceProvider.PollInstance(ctx, instance.GUID)
	if err != nil {
		return domain.LastOperation{State: domain.Failed, Description: err.Error()}, nil
	}

	if !done {
		return domain.LastOperation{State: domain.InProgress, Description: message}, nil
	}

	// the instance may have been invalidated, so we pass its primary key rather than the
	// instance directly.
	updateErr := broker.updateStateOnOperationCompletion(ctx, serviceProvider, lastOperationType, instanceID)

	return domain.LastOperation{State: domain.Succeeded, Description: message}, updateErr
}

// updateStateOnOperationCompletion handles updating/cleaning-up resources that need to be changed
// once lastOperation finishes successfully.
func (broker *ServiceBroker) updateStateOnOperationCompletion(ctx context.Context, service broker.ServiceProvider, lastOperationType, instanceID string) error {
	if lastOperationType == models.DeprovisionOperationType {
		if err := broker.store.DeleteServiceInstanceDetails(instanceID); err != nil {
			return fmt.Errorf("error deleting instance details from database: %s. WARNING: this instance will remain visible in cf. Contact your operator for cleanup", err)
		}
		if err := broker.store.DeleteProvisionRequestDetails(instanceID); err != nil {
			return fmt.Errorf("error deleting provision request details from the database: %w", err)
		}

		return nil
	}

	// If the operation was not a delete, clear out the ID and type and update
	// any changed (or finalized) state like IP addresses, selflinks, etc.
	details, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return fmt.Errorf("error getting instance details from database %v", err)
	}

	outs, err := service.GetTerraformOutputs(ctx, details.GUID)
	if err != nil {
		return fmt.Errorf("error getting new instance details: %s", err)
	}

	details.Outputs = outs
	details.OperationGUID = ""
	details.OperationType = models.ClearOperationType
	if err := broker.store.StoreServiceInstanceDetails(details); err != nil {
		return fmt.Errorf("error saving instance details to database %v", err)
	}

	return nil
}
