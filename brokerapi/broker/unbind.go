package broker

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager/v3"
	"github.com/pivotal-cf/brokerapi/v10/domain"
	"github.com/pivotal-cf/brokerapi/v10/domain/apiresponses"

	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
	"github.com/cloudfoundry/cloud-service-broker/utils/request"
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
	if err != nil {
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

	if broker.Credstore != nil {
		credentialName := getCredentialName(broker.getServiceName(serviceDefinition), bindingID)

		if err := broker.Credstore.DeletePermission(credentialName); err != nil {
			broker.Logger.Error(fmt.Sprintf("failed to delete permissions on the CredHub key %s", credentialName), err)
		}

		if err := broker.Credstore.Delete(credentialName); err != nil {
			broker.Logger.Error(fmt.Sprintf("failed to delete CredHub key %s", credentialName), err)
		}
	}

	// remove binding from database
	if err := broker.store.DeleteServiceBindingCredentials(bindingID, instanceID); err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error soft-deleting credentials from database: %s. WARNING: these credentials will remain visible in cf. Contact your operator for cleanup", err)
	}
	if err := broker.store.DeleteBindRequestDetails(bindingID, instanceID); err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error soft-deleting bind request details from database: %s", err)
	}

	if err := serviceProvider.DeleteBindingData(ctx, instanceID, bindingID); err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error deleting provider binding data from database: %s", err)
	}

	return domain.UnbindSpec{}, nil
}
