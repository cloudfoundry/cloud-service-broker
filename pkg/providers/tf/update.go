package tf

import (
	"context"

	"code.cloudfoundry.org/lager/v3"

	"github.com/cloudfoundry/cloud-service-broker/v3/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v3/utils/correlation"
)

// Update makes necessary updates to resources, so they match new desired configuration
func (provider *TerraformProvider) Update(ctx context.Context, updateContext *varcontext.VarContext) (models.ServiceInstanceDetails, error) {
	provider.logger.Debug("update", correlation.ID(ctx), lager.Data{
		"context": updateContext.ToMap(),
	})

	tfID := updateContext.GetString("tf_id")
	if err := updateContext.Error(); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	if err := provider.UpdateWorkspaceHCL(tfID, provider.serviceDefinition.ProvisionSettings, updateContext.ToMap()); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	deployment, err := provider.GetTerraformDeployment(tfID)
	if err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	workspace := deployment.Workspace

	if err := provider.MarkOperationStarted(&deployment, models.UpdateOperationType); err != nil {
		return models.ServiceInstanceDetails{}, err
	}

	go func() {
		err = workspace.UpdateInstanceConfiguration(updateContext.ToMap())
		if err != nil {
			_ = provider.MarkOperationFinished(&deployment, err)
			return
		}

		err = provider.DefaultInvoker().Apply(ctx, workspace)
		_ = provider.MarkOperationFinished(&deployment, err)
	}()

	return models.ServiceInstanceDetails{
		OperationID:   tfID,
		OperationType: models.UpdateOperationType,
	}, nil
}
