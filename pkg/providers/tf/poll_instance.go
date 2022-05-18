package tf

import (
	"context"
	"errors"
)

// PollInstance returns the instance status of the backing job.
func (provider *TerraformProvider) PollInstance(ctx context.Context, instanceGUID string) (bool, string, error) {
	deployment, err := provider.store.GetTerraformDeployment(generateTfID(instanceGUID, ""))
	if err != nil {
		return true, "", err
	}

	switch deployment.LastOperationState {
	case Succeeded:
		return true, deployment.LastOperationMessage, nil
	case Failed:
		return true, deployment.LastOperationMessage, errors.New(deployment.LastOperationMessage)
	default:
		return false, deployment.LastOperationMessage, nil
	}
}
