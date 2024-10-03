package tf

import (
	"context"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/correlation"
)

func (provider *TerraformProvider) ClearOperationType(ctx context.Context, instanceGUID string) error {
	provider.logger.Debug("ClearOperationType", correlation.ID(ctx), lager.Data{
		"instance": instanceGUID,
	})

	tfID := generateTfID(instanceGUID, "")

	return provider.ResetOperationType(tfID)
}
