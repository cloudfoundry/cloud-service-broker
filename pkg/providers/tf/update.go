package tf

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
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

	if err := provider.UpdateWorkspaceHCL(tfID, provider.serviceDefinition.ProvisionSettings, provisionContext.ToMap()); err != nil {
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
		err = provider.checkTerraformVersion(ctx, workspace)
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
