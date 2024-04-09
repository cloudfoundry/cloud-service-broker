package tf

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/correlation"
)

// Bind creates a new backing Terraform job and executes it, waiting on the result.
func (provider *TerraformProvider) Bind(ctx context.Context, bindContext *varcontext.VarContext) (map[string]any, error) {
	provider.logger.Debug("terraform-bind", correlation.ID(ctx), lager.Data{
		"context": bindContext.ToMap(),
	})

	tfID, err := provider.create(ctx, bindContext, provider.serviceDefinition.BindSettings, models.BindOperationType)
	if err != nil {
		return nil, fmt.Errorf("error from provider bind: %w", err)
	}

	if err := provider.Wait(ctx, tfID); err != nil {
		return nil, fmt.Errorf("error waiting for result: %w", err)
	}

	return provider.outputs(tfID, workspace.DefaultInstanceName)
}
