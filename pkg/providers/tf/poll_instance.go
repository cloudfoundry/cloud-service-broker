package tf

import (
	"context"
)

// PollInstance returns the instance status of the backing job.
func (provider *TerraformProvider) PollInstance(_ context.Context, instanceGUID string) (bool, string, string, error) {
	return provider.OperationStatus(generateTfID(instanceGUID, ""))
}
