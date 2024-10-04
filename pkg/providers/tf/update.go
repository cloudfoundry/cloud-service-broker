package tf

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager/v3"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/correlation"
)

// Update makes necessary updates to resources, so they match new desired configuration
func (provider *TerraformProvider) Update(ctx context.Context, updateContext *varcontext.VarContext) error {
	provider.logger.Debug("update", correlation.ID(ctx), lager.Data{
		"context": updateContext.ToMap(),
	})

	if provider.serviceDefinition.ProvisionSettings.IsTfImport(updateContext) {
		return fmt.Errorf("cannot update to subsume plan\n\nFor OpsMan Tile users see documentation here: https://via.vmw.com/ENs4\n\nFor Open Source users deployed via 'cf push' see documentation here:  https://via.vmw.com/ENw4")
	}

	tfID := updateContext.GetString("tf_id")
	if err := updateContext.Error(); err != nil {
		return err
	}

	if err := provider.UpdateWorkspaceHCL(tfID, provider.serviceDefinition.ProvisionSettings, updateContext.ToMap()); err != nil {
		return err
	}

	deployment, err := provider.GetTerraformDeployment(tfID)
	if err != nil {
		return err
	}

	workspace := deployment.Workspace

	if err := provider.MarkOperationStarted(&deployment, models.UpdateOperationType); err != nil {
		return err
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

	return nil
}
