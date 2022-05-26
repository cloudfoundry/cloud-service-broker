package tf

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

// Deprovision performs a terraform destroy on the instance.
func (provider *TerraformProvider) Deprovision(ctx context.Context, instanceGUID string, _ domain.DeprovisionDetails, vc *varcontext.VarContext) (*string, error) {
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
