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

// Upgrade makes necessary updates to resources so they match plan configuration
func (provider *TerraformProvider) Upgrade(ctx context.Context, upgradeContext *varcontext.VarContext) (models.ServiceInstanceDetails, error) {
	provider.logger.Debug("upgrade", correlation.ID(ctx), lager.Data{
		"context": upgradeContext.ToMap(),
	})

	tfID := upgradeContext.GetString("tf_id")
	if err := upgradeContext.Error(); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	if err := provider.UpdateWorkspaceHCL(tfID, provider.serviceDefinition.ProvisionSettings, upgradeContext.ToMap()); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	deployment, err := provider.GetTerraformDeployment(tfID)
	if err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	workspace := deployment.Workspace

	if err := provider.MarkOperationStarted(&deployment, models.UpgradeOperationType); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	go func() {
		err = provider.performTerraformUpgrade(ctx, workspace)
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
