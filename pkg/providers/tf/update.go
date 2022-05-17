package tf

import (
	"code.cloudfoundry.org/lager"
	"context"
	"fmt"
	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

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

	if err := UpdateWorkspaceHCL(provider.store, provider.serviceDefinition.ProvisionSettings, provisionContext, tfID); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	deployment, err := provider.store.GetTerraformDeployment(tfID)
	if err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	workspace := deployment.TFWorkspace()

	if err := provider.markJobStarted(deployment, models.UpdateOperationType); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	go func() {
		err = provider.performTerraformUpgrade(ctx, workspace)
		if err != nil {
			provider.operationFinished(err, deployment)
			return
		}

		err = workspace.UpdateInstanceConfiguration(provisionContext.ToMap())
		if err != nil {
			provider.operationFinished(err, deployment)
			return
		}

		err = provider.DefaultInvoker().Apply(ctx, workspace)
		provider.operationFinished(err, deployment)
	}()

	return models.ServiceInstanceDetails{
		OperationID:   tfID,
		OperationType: models.UpdateOperationType,
	}, nil
}
