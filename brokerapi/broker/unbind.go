package broker

import (
	"context"
	"errors"
	"fmt"

	"code.cloudfoundry.org/brokerapi/v13/domain"
	"code.cloudfoundry.org/brokerapi/v13/domain/apiresponses"
	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/paramparser"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/correlation"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/request"
)

// Unbind destroys an account and credentials with access to an instance of a service.
// It is bound to the `DELETE /v2/service_instances/:instance_id/service_bindings/:binding_id` endpoint and can be called using the `cf unbind-service` command.
func (broker *ServiceBroker) Unbind(ctx context.Context, instanceID, bindingID string, details domain.UnbindDetails, _ bool) (domain.UnbindSpec, error) {
	broker.Logger.Info("Unbinding", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
		"binding_id":  bindingID,
		"details":     details,
	})

	// verify the service exists and the plan exists
	serviceDefinition, serviceProvider, err := broker.getDefinitionAndProvider(details.ServiceID)
	if err != nil {
		return domain.UnbindSpec{}, err
	}

	// validate existence of binding
	exists, err := broker.store.ExistsServiceBindingCredentials(bindingID, instanceID)
	switch {
	case err != nil:
		return domain.UnbindSpec{}, fmt.Errorf("error locating service binding: %w", err)
	case !exists:
		return domain.UnbindSpec{}, apiresponses.ErrBindingDoesNotExist
	}

	// get existing service instance details
	instance, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		// in practice, we will not see this error as the binding will likely not exist
		return domain.UnbindSpec{}, apiresponses.ErrInstanceDoesNotExist
	}

	err = serviceProvider.CheckUpgradeAvailable(generateTFBindingID(instanceID, bindingID))
	switch {
	case errors.As(err, &workspace.CannotReadVersionError{}):
		// In the special case of not being able to read the version during unbind, we succeed immediately because:
		// - we have had feedback that failing here creates unwanted manual work for CF admins
		// - the Terraform state is empty or invalid, so it would be impossible for CSB to do a successful cleanup
		broker.Logger.Info("unbind-cannot-read-version", lager.Data{"error": err})
		if err := broker.removeBindingData(ctx, instanceID, bindingID, serviceDefinition, serviceProvider); err != nil {
			broker.Logger.Error("unbind-cleanup", err)
		}
		return domain.UnbindSpec{}, nil
	case err != nil:
		return domain.UnbindSpec{}, fmt.Errorf("failed to unbind: %s", err.Error())
	}

	plan, err := serviceDefinition.GetPlanByID(details.PlanID)
	if err != nil {
		return domain.UnbindSpec{}, err
	}

	storedParams, err := broker.store.GetBindRequestDetails(bindingID, instanceID)
	if err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error retrieving bind request details for %q: %w", instanceID, err)
	}

	parsedDetails := paramparser.BindDetails{
		PlanID:        details.PlanID,
		ServiceID:     details.ServiceID,
		RequestParams: storedParams,
	}

	vars, err := serviceDefinition.BindVariables(instance, bindingID, parsedDetails, plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return domain.UnbindSpec{}, err
	}

	// remove binding from service provider
	if err := serviceProvider.Unbind(ctx, instanceID, bindingID, vars); err != nil {
		return domain.UnbindSpec{}, err
	}

	if err := broker.removeBindingData(ctx, instanceID, bindingID, serviceDefinition, serviceProvider); err != nil {
		return domain.UnbindSpec{}, err
	}

	return domain.UnbindSpec{}, nil
}

func (broker *ServiceBroker) removeBindingData(ctx context.Context, instanceID, bindingID string, serviceDefinition *broker.ServiceDefinition, serviceProvider broker.ServiceProvider) error {
	// remove the credential from CredHub
	if err := broker.credStore.Delete(ctx, computeCredHubPath(broker.getServiceName(serviceDefinition), bindingID)); err != nil {
		broker.Logger.Error("errors removing credential from CredHub", err)
	}

	// remove binding from database
	if err := broker.store.DeleteServiceBindingCredentials(bindingID, instanceID); err != nil {
		return fmt.Errorf("error soft-deleting credentials from database: %s. WARNING: these credentials will remain visible in cf. Contact your operator for cleanup", err)
	}
	if err := broker.store.DeleteBindRequestDetails(bindingID, instanceID); err != nil {
		return fmt.Errorf("error soft-deleting bind request details from database: %s", err)
	}

	// delete the Terraform workspace from the database
	if err := serviceProvider.DeleteBindingData(ctx, instanceID, bindingID); err != nil {
		return fmt.Errorf("error deleting provider binding data from database: %s", err)
	}

	return nil
}
