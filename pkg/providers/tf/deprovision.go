package tf

import (
	"code.cloudfoundry.org/lager"
	"context"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

// Deprovision performs a terraform destroy on the instance.
func (provider *TerraformProvider) Deprovision(ctx context.Context, instanceGUID string, details domain.DeprovisionDetails, vc *varcontext.VarContext) (operationID *string, err error) {
	provider.logger.Debug("terraform-deprovision", correlation.ID(ctx), lager.Data{
		"instance": instanceGUID,
	})

	tfID := generateTfID(instanceGUID, "")

	if err := UpdateWorkspaceHCL(provider.store, provider.serviceDefinition.ProvisionSettings, vc, tfID); err != nil {
		return nil, err
	}

	if err := provider.Destroy(ctx, tfID, vc.ToMap()); err != nil {
		return nil, err
	}

	return &tfID, nil
}
