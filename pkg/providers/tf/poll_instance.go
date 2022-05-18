package tf

import (
	"context"
)

// PollInstance returns the instance Status of the backing job.
func (provider *TerraformProvider) PollInstance(ctx context.Context, instanceGUID string) (bool, string, error) {
	return provider.Status(generateTfID(instanceGUID, ""))
}
