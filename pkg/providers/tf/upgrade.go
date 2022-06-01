package tf

import (
	"context"
	"errors"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

// TODO:
// - Review tests for workspace
// - Understand UpdateWorkspaceHCL
// - Create binding upgrade context
// - Refactor
// - Add more test to integration test
// - Test tfID split

// Upgrade makes necessary updates to resources so they match plan configuration
func (provider *TerraformProvider) Upgrade(ctx context.Context, upgradeContext *varcontext.VarContext) (models.ServiceInstanceDetails, error) {
	provider.logger.Debug("upgrade", correlation.ID(ctx), lager.Data{
		"context": upgradeContext.ToMap(),
	})

	instanceDeploymentID := upgradeContext.GetString("tf_id")
	if err := upgradeContext.Error(); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	if err := provider.UpdateWorkspaceHCL(instanceDeploymentID, provider.serviceDefinition.ProvisionSettings, upgradeContext.ToMap()); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	deployment, err := provider.GetTerraformDeployment(instanceDeploymentID)
	if err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	workspace := deployment.Workspace

	if err := provider.MarkOperationStarted(&deployment, models.UpgradeOperationType); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	bindingDeployments, err := provider.GetBindingDeployments(instanceDeploymentID)
	//if err != nil {
	//	return models.ServiceInstanceDetails{}, err
	//}

	for _, bindingDeployment := range bindingDeployments {
		// getUpgradeContext
		//instance, err := broker.store.GetServiceInstanceDetails(instanceID)
		//if err != nil {
		//	return fmt.Errorf("error retrieving service instance details: %s", err)
		//}
		//
		//storedParams, err := broker.store.GetBindRequestDetails(binding.BindingID, instanceID)
		//if err != nil {
		//	return fmt.Errorf("error retrieving bind request details for %q: %w", instanceID, err)
		//}
		//
		//parsedDetails := paramparser.BindDetails{
		//	PlanID:        details.PlanID,
		//	ServiceID:     details.ServiceID,
		//	RequestParams: storedParams,
		//}
		//
		//vars, err := brokerService.BindVariables(instance, binding.BindingID, parsedDetails, plan, request.DecodeOriginatingIdentityHeader(ctx))
		//if err != nil {
		//	return err
		//}
		varContext, _ := varcontext.Builder().Build()
		if err := provider.UpdateWorkspaceHCL(bindingDeployment.ID, provider.serviceDefinition.ProvisionSettings, varContext.ToMap()); err != nil {
			return models.ServiceInstanceDetails{}, err
		}
	}

	bindingDeployments1, err := provider.GetBindingDeployments(instanceDeploymentID)
	//if err != nil {
	//	return models.ServiceInstanceDetails{}, err
	//}

	go func() {
		err = provider.performTerraformUpgrade(ctx, workspace)
		if err != nil {
			provider.MarkOperationFinished(&deployment, err)
			return
		}

		for _, bindingDeployment := range bindingDeployments1 {
			bindingWorkspace := bindingDeployment.Workspace
			err = provider.performTerraformUpgrade(ctx, bindingWorkspace)
			provider.MarkOperationFinished(&bindingDeployment, err)
			//if err != nil {
			//	provider.MarkOperationFinished(&deployment, err)
			//	return
			//}
		}

		provider.MarkOperationFinished(&deployment, err)
	}()

	return models.ServiceInstanceDetails{}, nil
}

func (provider *TerraformProvider) performTerraformUpgrade(ctx context.Context, workspace workspace.Workspace) error {
	currentTfVersion, err := workspace.StateVersion()
	if err != nil {
		return err
	}

	if currentTfVersion.LessThan(provider.tfBinContext.DefaultTfVersion) {
		if provider.tfBinContext.TfUpgradePath == nil || len(provider.tfBinContext.TfUpgradePath) == 0 {
			return errors.New("terraform version mismatch and no upgrade path specified")
		}
		for _, targetTfVersion := range provider.tfBinContext.TfUpgradePath {
			if currentTfVersion.LessThan(targetTfVersion) {
				err = provider.VersionedInvoker(targetTfVersion).Apply(ctx, workspace)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
