package broker

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/lager"
	"github.com/hashicorp/go-version"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"

	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/decider"
	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
	"github.com/cloudfoundry/cloud-service-broker/utils/request"
)

var ErrNonUpdatableParameter = apiresponses.NewFailureResponse(errors.New("attempt to update parameter that may result in service instance re-creation and data loss"), http.StatusBadRequest, "prohibited")

// Update a service instance plan.
// This functionality is not implemented and will return an error indicating that plan changes are not supported.
func (broker *ServiceBroker) Update(ctx context.Context, instanceID string, details domain.UpdateDetails, asyncAllowed bool) (domain.UpdateServiceSpec, error) {
	broker.Logger.Info("Updating", correlation.ID(ctx), lager.Data{
		"instance_id":        instanceID,
		"accepts_incomplete": asyncAllowed,
		"details":            details,
	})

	// make sure that instance actually exists
	exists, err := broker.store.ExistsServiceInstanceDetails(instanceID)
	switch {
	case err != nil:
		return domain.UpdateServiceSpec{}, fmt.Errorf("database error checking for existing instance: %s", err)
	case !exists:
		return domain.UpdateServiceSpec{}, apiresponses.ErrInstanceDoesNotExist
	}

	instance, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return domain.UpdateServiceSpec{}, fmt.Errorf("database error getting existing instance: %s", err)
	}

	serviceDefinition, serviceProvider, err := broker.getDefinitionAndProvider(instance.ServiceGUID)
	if err != nil {
		return domain.UpdateServiceSpec{}, err
	}

	parsedDetails, err := paramparser.ParseUpdateDetails(details)
	if err != nil {
		return domain.UpdateServiceSpec{}, ErrInvalidUserInput
	}

	// verify the service exists and the plan exists
	plan, err := serviceDefinition.GetPlanByID(parsedDetails.PlanID)
	if err != nil {
		return domain.UpdateServiceSpec{}, err
	}
	maintenanceInfoVersion, err := readMaintenanceInfoVersion(plan)
	if err != nil {
		return domain.UpdateServiceSpec{}, err
	}

	// verify async provisioning is allowed if it is required
	if !asyncAllowed {
		return domain.UpdateServiceSpec{}, apiresponses.ErrAsyncRequired
	}

	// Give the user a better error message if they give us a bad request
	if err := validateProvisionParameters(parsedDetails.RequestParams, serviceDefinition.ProvisionInputVariables, nil, plan); err != nil {
		return domain.UpdateServiceSpec{}, err
	}
	if !serviceDefinition.AllowedUpdate(parsedDetails.RequestParams) {
		return domain.UpdateServiceSpec{}, ErrNonUpdatableParameter
	}

	provisionDetails, err := broker.store.GetProvisionRequestDetails(instanceID)
	if err != nil {
		return domain.UpdateServiceSpec{}, fmt.Errorf("error retrieving provision request details for %q: %w", instanceID, err)
	}

	importedParams, err := serviceProvider.GetImportedProperties(ctx, instance.PlanGUID, instance.GUID, serviceDefinition.ProvisionInputVariables)
	if err != nil {
		return domain.UpdateServiceSpec{}, fmt.Errorf("error retrieving subsume parameters for %q: %w", instanceID, err)
	}

	mergedDetails, err := mergeJSON(provisionDetails, parsedDetails.RequestParams, importedParams)
	if err != nil {
		return domain.UpdateServiceSpec{}, fmt.Errorf("error merging update and provision details: %w", err)
	}

	vars, err := serviceDefinition.UpdateVariables(instanceID, parsedDetails, mergedDetails, *plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return domain.UpdateServiceSpec{}, err
	}

	operation, err := decider.DecideOperation(maintenanceInfoVersion, parsedDetails)
	switch {
	case err != nil:
		return domain.UpdateServiceSpec{}, fmt.Errorf("error deciding update path: %w", err)
	case operation == decider.Upgrade:
		return broker.doUpgrade(ctx, serviceDefinition, serviceProvider, instance, vars, plan)
	default:
		return broker.doUpdate(ctx, serviceProvider, instance, vars, parsedDetails, mergedDetails)
	}
}

func (broker *ServiceBroker) doUpgrade(ctx context.Context, serviceDefinition *broker.ServiceDefinition, serviceProvider broker.ServiceProvider, instance storage.ServiceInstanceDetails, instanceVars *varcontext.VarContext, plan *broker.ServicePlan) (domain.UpdateServiceSpec, error) {
	instanceUpgradeFinished, err := serviceProvider.UpgradeInstance(ctx, instanceVars)
	if err != nil {
		return domain.UpdateServiceSpec{}, err
	}

	go func() {
		instanceUpgradeFinished.Wait()

		deployment, err := broker.store.GetTerraformDeployment(generateTFInstanceID(instance.GUID))
		if err != nil {
			broker.storeUpgradeError(err, instance.GUID)
			return
		}

		// If the instance upgrade has failed, the error will already have been recorded,
		// so all we need to do is skip the upgrade of the bindings
		if deployment.LastOperationState != tf.InProgress {
			return
		}

		outs, err := serviceProvider.GetTerraformOutputs(ctx, instance.GUID)
		if err != nil {
			broker.storeUpgradeError(err, instance.GUID)
			return
		}

		instance.Outputs = outs
		if err := broker.store.StoreServiceInstanceDetails(instance); err != nil {
			broker.storeUpgradeError(err, instance.GUID)
			return
		}

		bindingContexts, err := broker.createAllBindingContexts(ctx, serviceDefinition, instance, plan)
		if err != nil {
			broker.storeUpgradeError(err, instance.GUID)
			return
		}

		err = serviceProvider.UpgradeBindings(ctx, instanceVars, bindingContexts)
		if err != nil {
			broker.storeUpgradeError(err, instance.GUID)
			return
		}
	}()

	return domain.UpdateServiceSpec{
		IsAsync:       true,
		DashboardURL:  "",
		OperationData: generateTFInstanceID(instance.GUID),
	}, nil
}

func (broker *ServiceBroker) doUpdate(ctx context.Context, serviceProvider broker.ServiceProvider, instance storage.ServiceInstanceDetails, vars *varcontext.VarContext, parsedDetails paramparser.UpdateDetails, mergedDetails map[string]any) (domain.UpdateServiceSpec, error) {
	err := serviceProvider.CheckUpgradeAvailable(generateTFInstanceID(instance.GUID))
	if err != nil {
		return domain.UpdateServiceSpec{}, fmt.Errorf("terraform version check failed: %s", err.Error())
	}

	instanceDetails, err := serviceProvider.Update(ctx, vars)
	if err != nil {
		return domain.UpdateServiceSpec{}, err
	}

	// save instance plan change
	if instance.PlanGUID != parsedDetails.PlanID {
		instance.PlanGUID = parsedDetails.PlanID
	}

	instance.OperationType = instanceDetails.OperationType
	instance.OperationGUID = instanceDetails.OperationID
	if err := broker.store.StoreServiceInstanceDetails(instance); err != nil {
		return domain.UpdateServiceSpec{}, fmt.Errorf("error saving instance details to database: %s. WARNING: this instance cannot be deprovisioned through cf. Contact your operator for cleanup", err)
	}

	// save provision request details
	if err = broker.store.StoreProvisionRequestDetails(instance.GUID, mergedDetails); err != nil {
		return domain.UpdateServiceSpec{}, fmt.Errorf("error saving provision request details to database: %s. Services relying on async provisioning will not be able to complete provisioning", err)
	}

	return domain.UpdateServiceSpec{
		IsAsync:       true,
		DashboardURL:  "",
		OperationData: instanceDetails.OperationID,
	}, nil
}

func (broker *ServiceBroker) createAllBindingContexts(ctx context.Context, serviceDefinition *broker.ServiceDefinition, instance storage.ServiceInstanceDetails, plan *broker.ServicePlan) ([]*varcontext.VarContext, error) {
	bindingIDs, err := broker.store.GetServiceBindingIDsForServiceInstance(instance.GUID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving binding for instance %q: %w", instance.GUID, err)
	}

	var bindingContexts []*varcontext.VarContext
	for _, bindingID := range bindingIDs {
		storedParams, err := broker.store.GetBindRequestDetails(bindingID, instance.GUID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving bind request details for instance %q: %w", instance.GUID, err)
		}

		parsedDetails := paramparser.BindDetails{
			PlanID:        instance.PlanGUID,
			ServiceID:     instance.ServiceGUID,
			RequestParams: storedParams,
		}
		vars, err := serviceDefinition.BindVariables(instance, bindingID, parsedDetails, plan, request.DecodeOriginatingIdentityHeader(ctx))
		if err != nil {
			return nil, fmt.Errorf("error constructing bind variables for instance %q: %w", instance.GUID, err)
		}
		bindingContexts = append(bindingContexts, vars)
	}
	return bindingContexts, nil
}

func (broker *ServiceBroker) storeUpgradeError(errorToStore error, instanceID string) {
	deployment, err := broker.store.GetTerraformDeployment(generateTFInstanceID(instanceID))
	if err != nil {
		broker.Logger.Error("error-storing-error-get", err)
		return
	}

	deployment.LastOperationState = tf.Failed
	deployment.LastOperationMessage = fmt.Sprintf("%s %s: %s", deployment.LastOperationType, tf.Failed, errorToStore)

	if err := broker.store.StoreTerraformDeployment(deployment); err != nil {
		broker.Logger.Error("error-storing-error-store", err)
	}
}

func mergeJSON(previousParams, newParams, importParams map[string]any) (map[string]any, error) {
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

func readMaintenanceInfoVersion(plan *broker.ServicePlan) (*version.Version, error) {
	if plan.MaintenanceInfo != nil && len(plan.MaintenanceInfo.Version) != 0 {
		maintenanceInfoVersion, err := version.NewVersion(plan.MaintenanceInfo.Version)
		if err != nil {
			return nil, fmt.Errorf("error parsing plan maintenance info version: %w", err)
		}
		return maintenanceInfoVersion, nil
	}
	return nil, nil
}
