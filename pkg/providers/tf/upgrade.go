package tf

import (
	"context"
	"errors"
	"sync"

	"code.cloudfoundry.org/lager/v3"
	"github.com/hashicorp/go-version"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/correlation"
)

func (provider *TerraformProvider) UpgradeInstance(ctx context.Context, instanceContext *varcontext.VarContext) (*sync.WaitGroup, error) {
	provider.logger.Debug("upgrade-instance", correlation.ID(ctx), lager.Data{
		"context": instanceContext.ToMap(),
	})

	instanceDeploymentID := instanceContext.GetString("tf_id")
	if err := instanceContext.Error(); err != nil {
		return nil, err
	}

	if err := provider.UpdateWorkspaceHCL(instanceDeploymentID, provider.serviceDefinition.ProvisionSettings, instanceContext.ToMap()); err != nil {
		return nil, err
	}

	instanceDeployment, err := provider.GetTerraformDeployment(instanceDeploymentID)
	if err != nil {
		return nil, err
	}

	if err := provider.MarkOperationStarted(&instanceDeployment, models.UpgradeOperationType); err != nil {
		return nil, err
	}

	var finished sync.WaitGroup

	finished.Go(func() {
		err = provider.performTerraformUpgrade(ctx, instanceDeployment.Workspace)
		if err != nil {
			_ = provider.MarkOperationFinished(&instanceDeployment, err)
			return
		}

		if err := provider.MarkOperationStarted(&instanceDeployment, models.UpgradeOperationType); err != nil {
			panic(err)
		}
	})

	return &finished, nil
}

func (provider *TerraformProvider) UpgradeBindings(ctx context.Context, instanceContext *varcontext.VarContext, bindingContexts []*varcontext.VarContext) error {
	provider.logger.Debug("upgrade-bindings", correlation.ID(ctx), lager.Data{
		"context": instanceContext.ToMap(),
	})

	instanceDeploymentID := instanceContext.GetString("tf_id")
	if err := instanceContext.Error(); err != nil {
		return err
	}

	instanceDeployment, err := provider.GetTerraformDeployment(instanceDeploymentID)
	if err != nil {
		return err
	}

	for _, bindingContext := range bindingContexts {
		bindingDeploymentID := bindingContext.GetString("tf_id")
		if err := provider.UpdateWorkspaceHCL(bindingDeploymentID, provider.serviceDefinition.BindSettings, bindingContext.ToMap()); err != nil {
			return err
		}
	}
	bindingDeployments, err := provider.GetBindingDeployments(instanceDeploymentID)
	if err != nil {
		return err
	}

	go func() {
		for i := range bindingDeployments {
			err = provider.performTerraformUpgrade(ctx, bindingDeployments[i].Workspace)
			_ = provider.MarkOperationFinished(&bindingDeployments[i], err)
			if err != nil {
				_ = provider.MarkOperationFinished(&instanceDeployment, err)
				return
			}
		}

		_ = provider.MarkOperationFinished(&instanceDeployment, err)
	}()

	return nil
}

func (provider *TerraformProvider) performTerraformUpgrade(ctx context.Context, workspace workspace.Workspace) error {
	currentTfVersion, err := workspace.StateTFVersion()
	if err != nil {
		return err
	}

	// Because an upgrade can fail, and when it fails we still record the higher TF version in the state, we
	// need to perform the upgrade to the current TF version before performing the upgrade on higher versions.
	// This allows failures to be re-tried. We could add code to try and keep track of the failures, but
	// because TF is idempotent, it's cleaner to just run Apply at the current version, which will be a
	// no-op for success cases.
	if currentTfVersion.LessThan(version.Must(version.NewVersion("1.5.0"))) {
		return errors.New("upgrade only supported for Terraform versions >= 1.5.0")
	} else if currentTfVersion.LessThanOrEqual(provider.tfBinContext.DefaultTfVersion) {
		if len(provider.tfBinContext.TfUpgradePath) == 0 {
			return errors.New("tofu version mismatch and no upgrade path specified")
		}
		for _, targetTfVersion := range provider.tfBinContext.TfUpgradePath {
			if currentTfVersion.LessThanOrEqual(targetTfVersion) {
				err = provider.VersionedInvoker(targetTfVersion).Apply(ctx, workspace)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
