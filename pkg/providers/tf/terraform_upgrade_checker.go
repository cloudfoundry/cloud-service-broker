package tf

import (
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
)

func (provider *TerraformProvider) CheckUpgradeAvailable(deploymentGUID string) error {
	deployment, err := provider.GetTerraformDeployment(deploymentGUID)
	if err != nil {
		return err
	}

	tfWorkspace := deployment.TFWorkspace()

	err = provider.checkTerraformVersion(tfWorkspace)
	if err != nil {
		return err
	}

	return nil
}

func (provider *TerraformProvider) checkTerraformVersion(workspace workspace.Workspace) error {
	currentTfVersion, err := workspace.StateTFVersion()
	if err != nil {
		return err
	}
	if currentTfVersion.LessThan(provider.tfBinContext.DefaultTfVersion) {
		return errors.New("operation attempted with newer version of OpenTofu than current state, upgrade the service before retrying operation")
	}
	return nil
}
