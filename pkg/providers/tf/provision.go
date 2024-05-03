package tf

import (
	"context"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v3/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v3/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v3/utils/correlation"
)

// Provision creates the necessary resources that an instance of this service
// needs to operate.
func (provider *TerraformProvider) Provision(ctx context.Context, provisionContext *varcontext.VarContext) (storage.ServiceInstanceDetails, error) {
	provider.logger.Debug("terraform-provision", correlation.ID(ctx), lager.Data{
		"context": provisionContext.ToMap(),
	})

	tfID, err := provider.create(ctx, provisionContext, provider.serviceDefinition.ProvisionSettings, models.ProvisionOperationType)
	if err != nil {
		return storage.ServiceInstanceDetails{}, err
	}

	return storage.ServiceInstanceDetails{
		OperationGUID: tfID,
		OperationType: models.ProvisionOperationType,
	}, nil
}
