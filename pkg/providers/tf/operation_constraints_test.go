package tf

import (
	"testing"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/invoker"
)

func TestTerraformProvider_CheckOperationConstraints(t *testing.T) {
	type fields struct {
		tfBinContext               executor.TFBinariesContext
		TerraformInvokerBuilder    invoker.TerraformInvokerBuilder
		logger                     lager.Logger
		serviceDefinition          TfServiceDefinitionV1
		DeploymentManagerInterface DeploymentManagerInterface
	}
	type args struct {
		deploymentID  string
		operationType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &TerraformProvider{
				tfBinContext:               tt.fields.tfBinContext,
				TerraformInvokerBuilder:    tt.fields.TerraformInvokerBuilder,
				logger:                     tt.fields.logger,
				serviceDefinition:          tt.fields.serviceDefinition,
				DeploymentManagerInterface: tt.fields.DeploymentManagerInterface,
			}
			if err := provider.CheckOperationConstraints(tt.args.deploymentID, tt.args.operationType); (err != nil) != tt.wantErr {
				t.Errorf("CheckOperationConstraints() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
