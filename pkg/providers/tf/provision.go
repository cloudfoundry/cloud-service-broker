package tf

import (
	"code.cloudfoundry.org/lager"
	"context"
	"fmt"
	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

func (provider *TerraformProvider) Provision(ctx context.Context, provisionContext *varcontext.VarContext) (storage.ServiceInstanceDetails, error) {
	provider.logger.Debug("terraform-provision", correlation.ID(ctx), lager.Data{
		"context": provisionContext.ToMap(),
	})

	var (
		tfID string
		err  error
	)

	switch provider.serviceDefinition.ProvisionSettings.IsTfImport(provisionContext) {
	case true:
		tfID, err = provider.importCreate(ctx, provisionContext, provider.serviceDefinition.ProvisionSettings)
	default:
		tfID, err = provider.create(ctx, provisionContext, provider.serviceDefinition.ProvisionSettings)
	}
	if err != nil {
		return storage.ServiceInstanceDetails{}, err
	}

	return storage.ServiceInstanceDetails{
		OperationGUID: tfID,
		OperationType: models.ProvisionOperationType,
	}, nil
}

func (provider *TerraformProvider) create(ctx context.Context, vars *varcontext.VarContext, action TfServiceDefinitionV1Action) (string, error) {
	tfID := vars.GetString("tf_id")
	if err := vars.Error(); err != nil {
		return "", err
	}

	workspace, err := workspace.NewWorkspace(vars.ToMap(), action.Template, action.Templates, []workspace.ParameterMapping{}, []string{}, []workspace.ParameterMapping{})
	if err != nil {
		return tfID, fmt.Errorf("error creating workspace: %w", err)
	}

	deployment, err := provider.createAndSaveDeployment(tfID, workspace)
	if err != nil {
		provider.logger.Error("terraform provider create failed", err)
		return tfID, fmt.Errorf("terraform provider create failed: %w", err)
	}

	if err := provider.markJobStarted(deployment, models.ProvisionOperationType); err != nil {
		return tfID, fmt.Errorf("error marking job started: %w", err)
	}

	go func() {
		err := provider.DefaultInvoker().Apply(ctx, workspace)
		provider.operationFinished(err, deployment)
	}()

	return tfID, nil
}

func (provider *TerraformProvider) importCreate(ctx context.Context, vars *varcontext.VarContext, action TfServiceDefinitionV1Action) (string, error) {
	varsMap := vars.ToMap()

	var parameterMappings, addParams []workspace.ParameterMapping

	for _, importParameterMapping := range action.ImportParameterMappings {
		parameterMappings = append(parameterMappings, workspace.ParameterMapping{
			TfVariable:    importParameterMapping.TfVariable,
			ParameterName: importParameterMapping.ParameterName,
		})
	}

	for _, addParam := range action.ImportParametersToAdd {
		addParams = append(addParams, workspace.ParameterMapping{
			TfVariable:    addParam.TfVariable,
			ParameterName: addParam.ParameterName,
		})
	}

	var importParams []ImportResource

	for _, importParam := range action.ImportVariables {
		if param, ok := varsMap[importParam.Name]; ok {
			importParams = append(importParams, ImportResource{TfResource: importParam.TfResource, IaaSResource: fmt.Sprintf("%v", param)})
		}
	}

	if len(importParams) != len(action.ImportVariables) {
		importFields := action.ImportVariables[0].Name
		for i := 1; i < len(action.ImportVariables); i++ {
			importFields = fmt.Sprintf("%s, %s", importFields, action.ImportVariables[i].Name)
		}

		return "", fmt.Errorf("must provide values for all import parameters: %s", importFields)
	}

	tfID := vars.GetString("tf_id")
	if err := vars.Error(); err != nil {
		return "", err
	}

	workspace, err := workspace.NewWorkspace(varsMap, "", action.Templates, parameterMappings, action.ImportParametersToDelete, addParams)
	if err != nil {
		return tfID, err
	}

	deployment, err := provider.createAndSaveDeployment(tfID, workspace)
	if err != nil {
		provider.logger.Error("terraform provider create failed", err)
		return tfID, err
	}

	if err := provider.markJobStarted(deployment, models.ProvisionOperationType); err != nil {
		return tfID, err
	}
	invoker := provider.DefaultInvoker()

	go func() {
		logger := utils.NewLogger("Import").WithData(correlation.ID(ctx))
		resources := make(map[string]string)
		for _, resource := range importParams {
			resources[resource.TfResource] = resource.IaaSResource
		}

		if err := invoker.Import(ctx, workspace, resources); err != nil {
			logger.Error("Import Failed", err)
			provider.operationFinished(err, deployment)
			return
		}
		mainTf, err := invoker.Show(ctx, workspace)
		if err == nil {
			var tf string
			var parameterVals map[string]string
			tf, parameterVals, err = workspace.Transformer.ReplaceParametersInTf(workspace.Transformer.AddParametersInTf(workspace.Transformer.CleanTf(mainTf)))
			if err == nil {
				for pn, pv := range parameterVals {
					workspace.Instances[0].Configuration[pn] = pv
				}
				workspace.Modules[0].Definitions["main"] = tf

				logger.Info("new workspace", lager.Data{
					"workspace": workspace,
					"tf":        tf,
				})

				err = provider.terraformPlanToCheckNoResourcesDeleted(invoker, ctx, workspace, logger)
				if err != nil {
					logger.Error("plan failed", err)
				} else {
					err = invoker.Apply(ctx, workspace)
				}
			}
		}
		provider.operationFinished(err, deployment)
	}()

	return tfID, nil
}
