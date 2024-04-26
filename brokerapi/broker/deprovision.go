package broker

import (
	"context"
	"errors"
	"fmt"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v3/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v3/internal/paramparser"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/v3/utils/correlation"
	"github.com/cloudfoundry/cloud-service-broker/v3/utils/request"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"github.com/pivotal-cf/brokerapi/v11/domain/apiresponses"
)

// Deprovision destroys an existing instance of a service.
// It is bound to the `DELETE /v2/service_instances/:instance_id` endpoint and can be called using the `cf delete-service` command.
// If a deprovision is asynchronous, the returned DeprovisionServiceSpec will contain the operation ID for tracking its progress.
func (broker *ServiceBroker) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, clientSupportsAsync bool) (domain.DeprovisionServiceSpec, error) {
	broker.Logger.Info("Deprovisioning", correlation.ID(ctx), lager.Data{
		"instance_id":        instanceID,
		"accepts_incomplete": clientSupportsAsync,
		"details":            details,
	})

	if !clientSupportsAsync {
		return domain.DeprovisionServiceSpec{}, apiresponses.ErrAsyncRequired
	}

	// make sure that instance actually exists
	exists, err := broker.store.ExistsServiceInstanceDetails(instanceID)
	switch {
	case err != nil:
		return domain.DeprovisionServiceSpec{}, fmt.Errorf("database error checking for existing instance: %s", err)
	case !exists:
		return domain.DeprovisionServiceSpec{}, apiresponses.ErrInstanceDoesNotExist
	}

	instance, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return domain.DeprovisionServiceSpec{}, fmt.Errorf("database error getting existing instance: %s", err)
	}

	serviceDefinition, serviceProvider, err := broker.getDefinitionAndProvider(instance.ServiceGUID)
	if err != nil {
		return domain.DeprovisionServiceSpec{}, err
	}

	deploymentID := generateTFInstanceID(instanceID)

	if err := serviceProvider.CheckOperationConstraints(deploymentID, models.DeprovisionOperationType); err != nil {
		return domain.DeprovisionServiceSpec{}, err
	}

	err = serviceProvider.CheckUpgradeAvailable(deploymentID)
	switch {
	case errors.As(err, &workspace.CannotReadVersionError{}):
		// In the special case of not being able to read the version during deprovision, we succeed immediately because:
		// - we have had feedback that failing here creates unwanted manual work for CF admins
		// - the Terraform state is empty or invalid, so it would be impossible for CSB to do a successful cleanup
		broker.Logger.Info("deprovision-cannot-read-version")
		if err := broker.removeServiceInstanceData(ctx, instanceID, serviceProvider); err != nil {
			broker.Logger.Error("deprovision-cleanup", err)
		}
		return domain.DeprovisionServiceSpec{}, nil
	case err != nil:
		return domain.DeprovisionServiceSpec{}, fmt.Errorf("failed to delete: %s", err.Error())
	}

	// verify the service exists and the plan exists
	plan, err := serviceDefinition.GetPlanByID(details.PlanID)
	if err != nil {
		return domain.DeprovisionServiceSpec{}, err
	}

	parameters, err := broker.store.GetProvisionRequestDetails(instanceID)
	if err != nil {
		return domain.DeprovisionServiceSpec{}, fmt.Errorf("error retrieving provision request details for %q: %w", instanceID, err)
	}

	provisionDetails := paramparser.ProvisionDetails{
		ServiceID:     details.ServiceID,
		PlanID:        details.PlanID,
		RequestParams: parameters,
	}

	// validate parameters meet the service's schema and merge the user vars with
	// the plan's
	vars, err := serviceDefinition.ProvisionVariables(instanceID, provisionDetails, *plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return domain.DeprovisionServiceSpec{}, err
	}

	operationID, err := serviceProvider.Deprovision(ctx, instance.GUID, vars)
	if err != nil {
		return domain.DeprovisionServiceSpec{}, err
	}

	response := domain.DeprovisionServiceSpec{
		IsAsync:       true,
		OperationData: *operationID,
	}

	instance.OperationType = models.DeprovisionOperationType
	instance.OperationGUID = *operationID
	if err := broker.store.StoreServiceInstanceDetails(instance); err != nil {
		return response, fmt.Errorf("error saving instance details to database: %s. WARNING: this instance will remain visible in cf. Contact your operator for cleanup", err)
	}
	return response, nil
}

func (broker *ServiceBroker) removeServiceInstanceData(ctx context.Context, instanceID string, serviceProvider broker.ServiceProvider) error {
	if err := broker.store.DeleteServiceInstanceDetails(instanceID); err != nil {
		return fmt.Errorf("error deleting instance details from database: %s. WARNING: this instance will remain visible in cf. Contact your operator for cleanup", err)
	}
	if err := broker.store.DeleteProvisionRequestDetails(instanceID); err != nil {
		return fmt.Errorf("error deleting provision request details from the database: %w", err)
	}
	if err := serviceProvider.DeleteInstanceData(ctx, instanceID); err != nil {
		return fmt.Errorf("error deleting provider instance data from database: %s", err)
	}

	return nil
}
