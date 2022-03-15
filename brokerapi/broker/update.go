package broker

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
	"github.com/cloudfoundry/cloud-service-broker/utils/request"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)

var ErrNonUpdatableParameter = apiresponses.NewFailureResponse(errors.New("attempt to update parameter that may result in service instance re-creation and data loss"), http.StatusBadRequest, "prohibited")

// Update a service instance plan.
// This functionality is not implemented and will return an error indicating that plan changes are not supported.
func (broker *ServiceBroker) Update(ctx context.Context, instanceID string, details domain.UpdateDetails, asyncAllowed bool) (response domain.UpdateServiceSpec, err error) {
	broker.Logger.Info("Updating", correlation.ID(ctx), lager.Data{
		"instance_id":        instanceID,
		"accepts_incomplete": asyncAllowed,
		"details":            details,
	})

	// make sure that instance actually exists
	exists, err := broker.store.ExistsServiceInstanceDetails(instanceID)
	switch {
	case err != nil:
		return response, fmt.Errorf("database error checking for existing instance: %s", err)
	case !exists:
		return response, apiresponses.ErrInstanceDoesNotExist
	}

	instance, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return response, fmt.Errorf("database error getting existing instance: %s", err)
	}

	brokerService, serviceHelper, err := broker.getDefinitionAndProvider(instance.ServiceGUID)
	if err != nil {
		return response, err
	}

	parsedDetails, err := paramparser.ParseUpdateDetails(details)
	if err != nil {
		return domain.UpdateServiceSpec{}, ErrInvalidUserInput
	}

	// verify the service exists and the plan exists
	plan, err := brokerService.GetPlanById(parsedDetails.PlanID)
	if err != nil {
		return response, err
	}

	// verify async provisioning is allowed if it is required
	shouldProvisionAsync := serviceHelper.ProvisionsAsync()
	if shouldProvisionAsync && !asyncAllowed {
		return response, apiresponses.ErrAsyncRequired
	}

	// Give the user a better error message if they give us a bad request
	if err := validateProvisionParameters(parsedDetails.RequestParams, brokerService.ProvisionInputVariables, nil, plan); err != nil {
		return domain.UpdateServiceSpec{}, err
	}

	allowUpdate, err := brokerService.AllowedUpdate(parsedDetails.RequestParams)
	if err != nil {
		return response, err
	}
	if !allowUpdate {
		return response, ErrNonUpdatableParameter
	}

	provisionDetails, err := broker.store.GetProvisionRequestDetails(instanceID)
	if err != nil {
		return response, fmt.Errorf("error retrieving provision request details for %q: %w", instanceID, err)
	}

	importedParams, err := serviceHelper.GetImportedProperties(ctx, instance.PlanGUID, instance.GUID, brokerService.ProvisionInputVariables)
	if err != nil {
		return response, fmt.Errorf("error retrieving subsume parameters for %q: %w", instanceID, err)
	}

	mergedDetails, err := mergeJSON(provisionDetails, parsedDetails.RequestParams, importedParams)
	if err != nil {
		return response, fmt.Errorf("error merging update and provision details: %w", err)
	}

	vars, err := brokerService.UpdateVariables(instanceID, parsedDetails, mergedDetails, *plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return response, err
	}

	// get instance details
	newInstanceDetails, err := serviceHelper.Update(ctx, vars)
	if err != nil {
		return domain.UpdateServiceSpec{}, err
	}

	// save instance plan change
	if instance.PlanGUID != parsedDetails.PlanID {
		instance.PlanGUID = parsedDetails.PlanID
		if err := broker.store.StoreServiceInstanceDetails(instance); err != nil {
			return domain.UpdateServiceSpec{}, fmt.Errorf("error saving instance details to database: %s. WARNING: this instance cannot be deprovisioned through cf. Contact your operator for cleanup", err)
		}
	}

	// save provision request details
	if err = broker.store.StoreProvisionRequestDetails(instanceID, mergedDetails); err != nil {
		return domain.UpdateServiceSpec{}, fmt.Errorf("error saving provision request details to database: %s. Services relying on async provisioning will not be able to complete provisioning", err)
	}

	response.IsAsync = shouldProvisionAsync
	response.DashboardURL = ""
	response.OperationData = newInstanceDetails.OperationId

	return response, nil
}

func mergeJSON(previousParams, newParams, importParams map[string]interface{}) (map[string]interface{}, error) {
	vc, err := varcontext.Builder().
		MergeMap(previousParams).
		MergeMap(importParams).
		MergeMap(newParams).
		Build()
	if err != nil {
		return nil, err
	}

	return vc.ToMap(), nil
}
