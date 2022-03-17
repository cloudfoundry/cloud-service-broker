package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
	"github.com/cloudfoundry/cloud-service-broker/utils/request"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
	"github.com/spf13/viper"
)

var (
	invalidUserInputMsg = "User supplied parameters must be in the form of a valid JSON map."
	ErrInvalidUserInput = apiresponses.NewFailureResponse(errors.New(invalidUserInputMsg), http.StatusBadRequest, "parsing-user-request")
)

// Bind creates an account with credentials to access an instance of a service.
// It is bound to the `PUT /v2/service_instances/:instance_id/service_bindings/:binding_id` endpoint and can be called using the `cf bind-service` command.
func (broker *ServiceBroker) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, clientSupportsAsync bool) (domain.Binding, error) {
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

	parsedDetails, err := paramparser.ParseBindDetails(details)
	if err != nil {
		return domain.Binding{}, err
	}

	// verify the service exists and the plan exists
	plan, err := serviceDefinition.GetPlanById(details.PlanID)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error getting service plan: %w", err)
	}

	// Give the user a better error message if they give us a bad request
	if err := validateBindParameters(details.GetRawParameters(), serviceDefinition.BindInputVariables); err != nil {
		return domain.Binding{}, err
	}

	// validate parameters meet the service's schema and merge the plan's vars with
	// the user's
	vars, err := serviceDefinition.BindVariables(instanceRecord, bindingID, details, plan, request.DecodeOriginatingIdentityHeader(ctx))
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
		ServiceGUID:         details.ServiceID,
		Credentials:         credsDetails,
	}
	if err := broker.store.CreateServiceBindingCredentials(newCreds); err != nil {
		return domain.Binding{}, fmt.Errorf("error saving credentials to database: %w. WARNING: these credentials cannot be unbound through cf. Please contact your operator for cleanup",
			err)
	}

	bindRequest := storage.BindRequestDetails{
		ServiceInstanceGUID: instanceID,
		ServiceBindingGUID:  bindingID,
		RequestDetails:      details.RawParameters,
	}

	if err := broker.store.StoreBindRequestDetails(bindRequest); err != nil {
		return domain.Binding{}, fmt.Errorf("error saving bind request details to database: %s. Unbind operations will not be able to complete", err)
	}

	binding, err := serviceProvider.BuildInstanceCredentials(ctx, newCreds.Credentials, instanceRecord.Outputs)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error building credentials: %w", err)
	}

	if broker.Credstore != nil {
		credentialName := getCredentialName(broker.getServiceName(serviceDefinition), bindingID)

		_, err := broker.Credstore.Put(credentialName, binding.Credentials)
		if err != nil {
			return domain.Binding{}, fmt.Errorf("bind failure: unable to put credentials in Credstore: %w", err)
		}

		_, err = broker.Credstore.AddPermission(credentialName, "mtls-app:"+details.AppGUID, []string{"read"})
		if err != nil {
			return domain.Binding{}, fmt.Errorf("bind failure: unable to add Credstore permissions to app: %w", err)
		}

		binding.Credentials = map[string]interface{}{
			"credhub-ref": credentialName,
		}
	}

	return *binding, nil
}

func validateBindParameters(rawParams json.RawMessage, validUserInputFields []broker.BrokerVariable) error {
	if len(rawParams) == 0 {
		return nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(rawParams, &params); err != nil {
		return ErrInvalidUserInput
	}

	// As this is a new check we have feature-flagged it so that it can easily be disabled
	// if it causes problems.
	if !viper.GetBool(DisableRequestPropertyValidation) {
		return validateDefinedParams(params, validUserInputFields, nil)
	}
	return nil
}
