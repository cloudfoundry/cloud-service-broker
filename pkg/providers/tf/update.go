package tf

import (
	"context"
	"errors"
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
	"github.com/spf13/viper"
)

// Update makes necessary updates to resources so they match new desired configuration
func (provider *TerraformProvider) Update(ctx context.Context, provisionContext *varcontext.VarContext) (models.ServiceInstanceDetails, error) {
	provider.logger.Debug("update", correlation.ID(ctx), lager.Data{
		"context": provisionContext.ToMap(),
	})

	if provider.serviceDefinition.ProvisionSettings.IsTfImport(provisionContext) {
		return models.ServiceInstanceDetails{}, fmt.Errorf("cannot update to subsume plan\n\nFor OpsMan Tile users see documentation here: https://via.vmw.com/ENs4\n\nFor Open Source users deployed via 'cf push' see documentation here:  https://via.vmw.com/ENw4")
	}

	tfID := provisionContext.GetString("tf_id")
	if err := provisionContext.Error(); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	if err := provider.UpdateWorkspaceHCL(provider.serviceDefinition.ProvisionSettings, provisionContext, tfID); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	deployment, err := provider.GetTerraformDeployment(tfID)
	if err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	workspace := deployment.Workspace

	if err := provider.MarkOperationStarted(deployment, models.UpdateOperationType); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	go func() {
		err = provider.performTerraformUpgrade(ctx, workspace)
		if err != nil {
			provider.MarkOperationFinished(deployment, err)
			return
		}

		err = workspace.UpdateInstanceConfiguration(provisionContext.ToMap())
		if err != nil {
			provider.MarkOperationFinished(deployment, err)
			return
		}

		err = provider.DefaultInvoker().Apply(ctx, workspace)
		provider.MarkOperationFinished(deployment, err)
	}()

	return models.ServiceInstanceDetails{
		OperationID:   tfID,
		OperationType: models.UpdateOperationType,
	}, nil
}

func (provider *TerraformProvider) performTerraformUpgrade(ctx context.Context, workspace workspace.Workspace) error {
	currentTfVersion, err := workspace.StateVersion()
	if err != nil {
		return err
	}

	if viper.GetBool(TfUpgradeEnabled) {
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
	} else if currentTfVersion.LessThan(provider.tfBinContext.DefaultTfVersion) {
		return errors.New("apply attempted with a newer version of terraform than the state")
	}

	return nil
}
