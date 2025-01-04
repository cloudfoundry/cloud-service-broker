package broker

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/paramparser"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/featureflags"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/correlation"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/request"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"github.com/pivotal-cf/brokerapi/v11/domain/apiresponses"
)

// Bind creates an account with credentials to access an instance of a service.
// It is bound to the `PUT /v2/service_instances/:instance_id/service_bindings/:binding_id` endpoint and can be called using the `cf bind-service` command.
func (broker *ServiceBroker) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, _ bool) (domain.Binding, error) {
	broker.Logger.Info("Binding", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
		"binding_id":  bindingID,
		"details":     details,
	})

	// check for existing binding
	exists, err := broker.store.ExistsServiceBindingCredentials(bindingID, instanceID)
	switch {
	case err != nil:
		return domain.Binding{}, fmt.Errorf("error checking for existing binding: %w", err)
	case exists:
		return domain.Binding{}, apiresponses.ErrBindingAlreadyExists
	}

	// get existing service instance details
	instanceRecord, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error retrieving service instance details: %w", err)
	}

	serviceDefinition, serviceProvider, err := broker.getDefinitionAndProvider(instanceRecord.ServiceGUID)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error retrieving service definition: %w", err)
	}

	err = serviceProvider.CheckUpgradeAvailable(generateTFInstanceID(instanceID))
	if err != nil {
		return domain.Binding{}, fmt.Errorf("failed to bind: %s", err.Error())
	}

	parsedDetails, err := paramparser.ParseBindDetails(details)
	if err != nil {
		return domain.Binding{}, ErrInvalidUserInput
	}

	// verify the service exists and the plan exists
	plan, err := serviceDefinition.GetPlanByID(parsedDetails.PlanID)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error getting service plan: %w", err)
	}

	// Give the user a better error message if they give us a bad request
	if err := validateBindParameters(parsedDetails.RequestParams, serviceDefinition.BindInputVariables); err != nil {
		return domain.Binding{}, err
	}

	// validate parameters meet the service's schema and merge the plan's vars with
	// the user's
	vars, err := serviceDefinition.BindVariables(instanceRecord, bindingID, parsedDetails, plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error generating bind variables: %w", err)
	}

	// create binding
	credsDetails, err := serviceProvider.Bind(ctx, vars)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error performing bind: %w", err)
	}

	// save binding to database
	newCreds := storage.ServiceBindingCredentials{
		ServiceInstanceGUID: instanceID,
		BindingGUID:         bindingID,
		ServiceGUID:         parsedDetails.ServiceID,
		Credentials:         credsDetails,
	}
	if err := broker.store.CreateServiceBindingCredentials(newCreds); err != nil {
		return domain.Binding{}, fmt.Errorf("error saving credentials to database: %w. WARNING: these credentials cannot be unbound through cf. Please contact your operator for cleanup",
			err)
	}

	bindRequest := storage.BindRequestDetails{
		ServiceInstanceGUID: instanceID,
		ServiceBindingGUID:  bindingID,
		RequestDetails:      parsedDetails.RequestParams,
	}

	if err := broker.store.StoreBindRequestDetails(bindRequest); err != nil {
		return domain.Binding{}, fmt.Errorf("error saving bind request details to database: %s. Unbind operations will not be able to complete", err)
	}

	binding, err := buildInstanceCredentials(newCreds.Credentials, instanceRecord.Outputs)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error building credentials: %w", err)
	}

	if broker.Credstore != nil {
		credentialName := getCredentialName(broker.getServiceName(serviceDefinition), bindingID)

		_, err := broker.Credstore.Put(credentialName, binding.Credentials)
		if err != nil {
			return domain.Binding{}, fmt.Errorf("bind failure: unable to put credentials in Credstore: %w", err)
		}

		_, err = broker.Credstore.AddPermission(credentialName, "mtls-app:"+parsedDetails.AppGUID, []string{"read"})
		if err != nil {
			return domain.Binding{}, fmt.Errorf("bind failure: unable to add Credstore permissions to app: %w", err)
		}

		binding.Credentials = map[string]any{
			"credhub-ref": credentialName,
		}
	}

	return *binding, nil
}

func validateBindParameters(params map[string]any, validUserInputFields []broker.BrokerVariable) error {
	if len(params) == 0 {
		return nil
	}

	// As this is a new check we have feature-flagged it so that it can easily be disabled
	// if it causes problems.
	if !featureflags.Enabled(featureflags.DisableRequestPropertyValidation) {
		return validateDefinedParams(params, validUserInputFields, nil)
	}
	return nil
}

// BuildInstanceCredentials combines the bind credentials with the connection
// information in the instance details to get a full set of connection details.
func buildInstanceCredentials(credentials map[string]any, outputs storage.JSONObject) (*domain.Binding, error) {
	vc, err := varcontext.Builder().
		MergeMap(outputs).
		MergeMap(credentials).
		Build()
	if err != nil {
		return nil, err
	}

	return &domain.Binding{Credentials: vc.ToMap()}, nil
}
