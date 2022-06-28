package tf

import (
	"context"
	"errors"

	"github.com/hashicorp/go-version"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

// Upgrade makes necessary updates to resources so they match plan configuration
func (provider *TerraformProvider) Upgrade(ctx context.Context, instanceContext *varcontext.VarContext, bindingContexts []*varcontext.VarContext) (models.ServiceInstanceDetails, error) {
	provider.logger.Debug("upgrade", correlation.ID(ctx), lager.Data{
		"context": instanceContext.ToMap(),
	})

	instanceDeploymentID := instanceContext.GetString("tf_id")
	if err := instanceContext.Error(); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	if err := provider.UpdateWorkspaceHCL(instanceDeploymentID, provider.serviceDefinition.ProvisionSettings, instanceContext.ToMap()); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	instanceDeployment, err := provider.GetTerraformDeployment(instanceDeploymentID)
	if err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	if err := provider.MarkOperationStarted(&instanceDeployment, models.UpgradeOperationType); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	for _, bindingContext := range bindingContexts {
		bindingDeploymentID := bindingContext.GetString("tf_id")
		if err := provider.UpdateWorkspaceHCL(bindingDeploymentID, provider.serviceDefinition.BindSettings, bindingContext.ToMap()); err != nil {
			return models.ServiceInstanceDetails{}, err
		}
	}
	bindingDeployments, err := provider.GetBindingDeployments(instanceDeploymentID)
	if err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	go func() {
		err = provider.performTerraformUpgrade(ctx, instanceDeployment.Workspace)
		if err != nil {
			provider.MarkOperationFinished(&instanceDeployment, err)
			return
		}

		for i := range bindingDeployments {
			err = provider.performTerraformUpgrade(ctx, bindingDeployments[i].Workspace)
			provider.MarkOperationFinished(&bindingDeployments[i], err)
			if err != nil {
				provider.MarkOperationFinished(&instanceDeployment, err)
				return
			}
		}

		provider.MarkOperationFinished(&instanceDeployment, err)
	}()

	return models.ServiceInstanceDetails{}, nil
}

func (provider *TerraformProvider) performTerraformUpgrade(ctx context.Context, workspace workspace.Workspace) error {
	currentTfVersion, err := workspace.StateTFVersion()
	if err != nil {
		return err
	}

	if currentTfVersion.LessThan(version.Must(version.NewVersion("0.12.0"))) {
		return errors.New("upgrade only supported for Terraform versions >= 0.12.0")
	} else if currentTfVersion.LessThan(provider.tfBinContext.DefaultTfVersion) {
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
