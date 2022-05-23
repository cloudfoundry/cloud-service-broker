package tf

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/invoker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

// ImportResource represents TF resource to IaaS resource ID mapping for import
type ImportResource struct {
	TfResource   string
	IaaSResource string
}

// Provision creates the necessary resources that an instance of this service
// needs to operate.
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

func (provider *TerraformProvider) importCreate(ctx context.Context, vars *varcontext.VarContext, action TfServiceDefinitionV1Action) (string, error) {
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
	varsMap := vars.ToMap()
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
		return tfID, fmt.Errorf("error creating workspace: %w", err)
	}

	deployment, err := provider.CreateAndSaveDeployment(tfID, workspace)
	if err != nil {
		provider.logger.Error("terraform provider create failed", err)
		return tfID, fmt.Errorf("terraform provider create failed: %w", err)
	}

	if err := provider.MarkOperationStarted(deployment, models.ProvisionOperationType); err != nil {
		return tfID, fmt.Errorf("error marking job started: %w", err)
	}

	go func() {
		logger := utils.NewLogger("Import").WithData(correlation.ID(ctx))
		resources := make(map[string]string)
		for _, resource := range importParams {
			resources[resource.TfResource] = resource.IaaSResource
		}

		invoker := provider.DefaultInvoker()
		if err := invoker.Import(ctx, workspace, resources); err != nil {
			logger.Error("tf import failed", err)
			provider.MarkOperationFinished(deployment, fmt.Errorf("tf import failed: %w", err))
			return
		}

		mainTf, err := invoker.Show(ctx, workspace)
		if err != nil {
			logger.Error("tf show failed", err)
			provider.MarkOperationFinished(deployment, fmt.Errorf("tf show failed: %w", err))
		}

		err = createTFMainDefinition(workspace, mainTf, deployment, logger)
		if err != nil {
			logger.Error("Failed to create TF definition", err)
			provider.MarkOperationFinished(deployment, err)
		}

		err = provider.terraformPlanToCheckNoResourcesDeleted(invoker, ctx, workspace, logger)
		if err != nil {
			logger.Error("plan failed", err)
		} else {
			err = invoker.Apply(ctx, workspace)
		}

		provider.MarkOperationFinished(deployment, err)
	}()

	return tfID, nil
}

func createTFMainDefinition(workspace *workspace.TerraformWorkspace, mainTf string, deployment storage.TerraformDeployment, logger lager.Logger) error {
	var tf string
	var parameterVals map[string]string
	tf, parameterVals, err := workspace.Transformer.ReplaceParametersInTf(workspace.Transformer.AddParametersInTf(workspace.Transformer.CleanTf(mainTf)))
	if err != nil {
		return err
	}

	for pn, pv := range parameterVals {
		workspace.Instances[0].Configuration[pn] = pv
	}
	workspace.Modules[0].Definitions["main"] = tf

	logger.Info("new workspace", lager.Data{
		"workspace": workspace,
		"tf":        tf,
	})

	return nil
}

func (provider *TerraformProvider) terraformPlanToCheckNoResourcesDeleted(invoker invoker.TerraformInvoker, ctx context.Context, workspace *workspace.TerraformWorkspace, logger lager.Logger) error {
	planOutput, err := invoker.Plan(ctx, workspace)
	if err != nil {
		return err
	}
	err = CheckTerraformPlanOutput(logger, planOutput)
	return err
}
