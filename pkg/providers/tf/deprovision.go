package tf

import (
	"context"

	"code.cloudfoundry.org/lager"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
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
