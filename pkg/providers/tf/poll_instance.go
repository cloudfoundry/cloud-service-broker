package tf

import (
	"context"
)

// PollInstance returns the instance status of the backing job.
func (provider *TerraformProvider) PollInstance(ctx context.Context, instanceGUID string) (bool, string, error) {
	return provider.status(generateTfID(instanceGUID, ""))
}
