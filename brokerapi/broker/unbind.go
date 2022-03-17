package broker

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
	"github.com/cloudfoundry/cloud-service-broker/utils/request"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)

// Unbind destroys an account and credentials with access to an instance of a service.
// It is bound to the `DELETE /v2/service_instances/:instance_id/service_bindings/:binding_id` endpoint and can be called using the `cf unbind-service` command.
func (broker *ServiceBroker) Unbind(ctx context.Context, instanceID, bindingID string, details domain.UnbindDetails, asyncSupported bool) (domain.UnbindSpec, error) {
	broker.Logger.Info("Unbinding", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
		"binding_id":  bindingID,
		"details":     details,
	})

	serviceDefinition, serviceProvider, err := broker.getDefinitionAndProvider(details.ServiceID)
	if err != nil {
		return domain.UnbindSpec{}, err
	}

	parsedDetails, err := paramparser.ParseUnbindDetails(details)
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
		return domain.UnbindSpec{}, fmt.Errorf("error retrieving service instance details: %s", err)
	}

	// verify the service exists and the plan exists
	plan, err := serviceDefinition.GetPlanById(parsedDetails.PlanID)
	if err != nil {
		return domain.UnbindSpec{}, err
	}

	storedParams, err := broker.store.GetBindRequestDetails(bindingID, instanceID)
	if err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error retrieving bind request details for %q: %w", instanceID, err)
	}

	parsedDetails.RequestParams = storedParams

	vars, err := serviceDefinition.BindVariables(instance, bindingID, parsedDetails, plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return domain.UnbindSpec{}, err
	}

	// remove binding from service provider
	if err := serviceProvider.Unbind(ctx, instanceID, bindingID, vars); err != nil {
		return domain.UnbindSpec{}, err
	}

	if broker.Credstore != nil {
		credentialName := getCredentialName(broker.getServiceName(serviceDefinition), bindingID)

		err = broker.Credstore.DeletePermission(credentialName)
		if err != nil {
			broker.Logger.Error(fmt.Sprintf("fail to delete permissions on the key %s", credentialName), err)
		}

		err := broker.Credstore.Delete(credentialName)
		if err != nil {
			return domain.UnbindSpec{}, err
		}
	}

	// remove binding from database
	if err := broker.store.DeleteServiceBindingCredentials(bindingID, instanceID); err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error soft-deleting credentials from database: %s. WARNING: these credentials will remain visible in cf. Contact your operator for cleanup", err)
	}
	if err := broker.store.DeleteBindRequestDetails(bindingID, instanceID); err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error soft-deleting bind request details from database: %s", err)
	}

	return domain.UnbindSpec{}, nil
}
