package tf

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/spf13/viper"
)

const (
	dynamicHCLEnabled = "brokerpak.updates.enabled"
)

func init() {
	viper.BindEnv(dynamicHCLEnabled, "BROKERPAK_UPDATES_ENABLED")
	viper.SetDefault(dynamicHCLEnabled, false)
}

func UpdateWorkspaceHCL(store broker.ServiceProviderStorage, action TfServiceDefinitionV1Action, operationContext *varcontext.VarContext, tfID string) error {
	if !viper.GetBool(dynamicHCLEnabled) {
		return nil
	}
	deployment, err := store.GetTerraformDeployment(tfID)
	if err != nil {
		return err
	}

	currentWorkspace := deployment.TFWorkspace()
	if err != nil {
		return err
	}

	workspace, err := workspace.NewWorkspace(operationContext.ToMap(), action.Template, action.Templates, []workspace.ParameterMapping{}, []string{}, []workspace.ParameterMapping{})
	if err != nil {
		return err
	}

	workspace.State = currentWorkspace.State

	deployment.Workspace = workspace
	if err := store.StoreTerraformDeployment(deployment); err != nil {
		return fmt.Errorf("terraform provider create failed: %w", err)
	}

	return nil
}
