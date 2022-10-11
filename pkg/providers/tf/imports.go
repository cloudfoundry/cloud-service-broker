package tf

import (
	"context"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/hclparser"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

func (provider *TerraformProvider) GetImportedProperties(ctx context.Context, instanceGUID string, inputVariables []broker.BrokerVariable, initialProperties map[string]any) (map[string]any, error) {
	provider.logger.Debug("getImportedProperties", correlation.ID(ctx), lager.Data{})

	varsToReplace, err := getVarsToReplace(inputVariables, initialProperties)
	switch {
	case err != nil:
		return nil, err
	case len(varsToReplace) == 0:
		return map[string]any{}, nil
	}

	deployment, err := provider.GetTerraformDeployment(generateTfID(instanceGUID, ""))
	if err != nil {
		return nil, err
	}

	tfHCL, err := provider.DefaultInvoker().Show(ctx, deployment.Workspace)
	if err != nil {
		return map[string]any{}, err
	}

	return hclparser.GetParameters(tfHCL, varsToReplace)
}

func getVarsToReplace(inputVariables []broker.BrokerVariable, initialProperties map[string]any) ([]hclparser.ExtractVariable, error) {
	var varsToReplace []hclparser.ExtractVariable
	for _, vars := range inputVariables {
		if vars.TFAttribute == "" {
			continue
		}

		if vars.TFAttributeSkip != "" {
			// When we try to evaluate TFAttributeSkip we haven't computed the
			// whole varcontext, so some fields may be missing their default
			// values. When the attribute can't be looked up, we act as if
			// it was set to "false". This is because we don't want an error
			// when the value of the field would simply default to false if
			// the whole varcontext had been computed. This does mean that
			// a reference to an invalid field won't fail, but that's better
			// than introducing unexpected errors.
			if skip, _ := initialProperties[vars.TFAttributeSkip].(bool); skip {
				continue
			}
		}

		varsToReplace = append(varsToReplace, hclparser.ExtractVariable{
			FieldToRead:  vars.TFAttribute,
			FieldToWrite: vars.FieldName,
		})
	}
	return varsToReplace, nil
}
