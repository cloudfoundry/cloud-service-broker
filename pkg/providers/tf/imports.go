package tf

import (
	"code.cloudfoundry.org/lager"
	"context"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/hclparser"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

func (provider *TerraformProvider) GetImportedProperties(ctx context.Context, planGUID string, instanceGUID string, inputVariables []broker.BrokerVariable) (map[string]interface{}, error) {
	provider.logger.Debug("getImportedProperties", correlation.ID(ctx), lager.Data{})

	if provider.serviceDefinition.IsSubsumePlan(planGUID) {
		return map[string]interface{}{}, nil
	}

	varsToReplace := getVarsToReplace(inputVariables)
	if len(varsToReplace) == 0 {
		return map[string]interface{}{}, nil
	}

	deployment, err := provider.GetTerraformDeployment(generateTfID(instanceGUID, ""))
	if err != nil {
		return nil, err
	}

	tfHCL, err := provider.DefaultInvoker().Show(ctx, deployment.Workspace)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return hclparser.GetParameters(tfHCL, varsToReplace)
}

func getVarsToReplace(inputVariables []broker.BrokerVariable) []hclparser.ExtractVariable {
	var varsToReplace []hclparser.ExtractVariable
	for _, vars := range inputVariables {
		if vars.TFAttribute != "" {
			varsToReplace = append(varsToReplace, hclparser.ExtractVariable{
				FieldToRead:  vars.TFAttribute,
				FieldToWrite: vars.FieldName,
			})
		}
	}
	return varsToReplace
}
