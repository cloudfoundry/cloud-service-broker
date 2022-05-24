package tf

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
)

func (provider *TerraformProvider) GetTerraformOutputs(ctx context.Context, instanceGUID string) (storage.JSONObject, error) {
	tfID := generateTfID(instanceGUID, "")

	outs, err := provider.outputs(tfID, workspace.DefaultInstanceName)
	if err != nil {
		return nil, err
	}

	return outs, nil
}

// Outputs gets the output variables for the given module instance in the workspace.
func (provider *TerraformProvider) outputs(deploymentID, instanceName string) (map[string]interface{}, error) {
	deployment, err := provider.GetTerraformDeployment(deploymentID)
	if err != nil {
		return nil, fmt.Errorf("error getting TF deployment: %w", err)
	}

	return deployment.Workspace.Outputs(instanceName)
}
