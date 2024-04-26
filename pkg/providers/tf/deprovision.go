package tf

import (
	"context"

	"code.cloudfoundry.org/lager/v3"

	"github.com/cloudfoundry/cloud-service-broker/v3/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v3/utils/correlation"
)

// Deprovision performs a terraform destroy on the instance.
func (provider *TerraformProvider) Deprovision(ctx context.Context, instanceGUID string, vc *varcontext.VarContext) (*string, error) {
	provider.logger.Debug("terraform-deprovision", correlation.ID(ctx), lager.Data{
		"instance": instanceGUID,
	})

	tfID := generateTfID(instanceGUID, "")

	if err := provider.UpdateWorkspaceHCL(tfID, provider.serviceDefinition.ProvisionSettings, vc.ToMap()); err != nil {
		return nil, err
	}

	if err := provider.destroy(ctx, tfID, vc.ToMap(), models.DeprovisionOperationType); err != nil {
		return nil, err
	}
	return &tfID, nil
}

// DeleteInstanceData deletes a terraform deployment from the database
func (provider *TerraformProvider) DeleteInstanceData(ctx context.Context, instanceGUID string) error {
	provider.logger.Debug("terraform-delete-instance-data", correlation.ID(ctx), lager.Data{
		"instance": instanceGUID,
	})

	tfID := generateTfID(instanceGUID, "")

	if err := provider.DeleteTerraformDeployment(tfID); err != nil {
		return err
	}

	return nil
}
