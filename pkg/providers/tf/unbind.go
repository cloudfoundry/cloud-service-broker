package tf

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/v3/dbservice/models"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v3/utils/correlation"
)

// Unbind performs a terraform destroy on the binding.
func (provider *TerraformProvider) Unbind(ctx context.Context, instanceGUID, bindingID string, vc *varcontext.VarContext) error {
	tfID := generateTfID(instanceGUID, bindingID)
	provider.logger.Debug("terraform-unbind", correlation.ID(ctx), lager.Data{
		"instance": instanceGUID,
		"binding":  bindingID,
		"tfId":     tfID,
	})

	if err := provider.UpdateWorkspaceHCL(tfID, provider.serviceDefinition.BindSettings, vc.ToMap()); err != nil {
		return err
	}

	if err := provider.destroy(ctx, tfID, vc.ToMap(), models.UnbindOperationType); err != nil {
		return err
	}

	return provider.Wait(ctx, tfID)
}

// DeleteBindingData deletes a terraform deployment from the database
func (provider *TerraformProvider) DeleteBindingData(ctx context.Context, instanceGUID, bindingID string) error {
	tfID := generateTfID(instanceGUID, bindingID)
	provider.logger.Debug("terraform-delete-binding-data", correlation.ID(ctx), lager.Data{
		"instance": instanceGUID,
		"binding":  bindingID,
		"tfId":     tfID,
	})

	if err := provider.DeleteTerraformDeployment(tfID); err != nil {
		return err
	}

	return nil
}
