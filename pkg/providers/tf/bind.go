package tf

import (
	"code.cloudfoundry.org/lager"
	"context"
	"fmt"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

// Bind creates a new backing Terraform job and executes it, waiting on the result.
func (provider *TerraformProvider) Bind(ctx context.Context, bindContext *varcontext.VarContext) (map[string]interface{}, error) {
	provider.logger.Debug("terraform-bind", correlation.ID(ctx), lager.Data{
		"context": bindContext.ToMap(),
	})

	tfID, err := provider.create(ctx, bindContext, provider.serviceDefinition.BindSettings)
	if err != nil {
		return nil, fmt.Errorf("error from provider bind: %w", err)
	}

	if err := provider.Wait(ctx, tfID); err != nil {
		return nil, fmt.Errorf("error from job runner: %w", err)
	}

	return provider.Outputs(ctx, tfID, workspace.DefaultInstanceName)
}
