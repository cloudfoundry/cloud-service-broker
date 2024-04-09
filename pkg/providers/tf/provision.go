package tf

import (
	"context"
	"fmt"
	"strings"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/steps"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
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
		tfID, err = provider.create(ctx, provisionContext, provider.serviceDefinition.ProvisionSettings, models.ProvisionOperationType)
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
	varsMap := vars.ToMap()
	importParams, err := validateImportParams(action.ImportVariables, varsMap)
	if err != nil {
		return "", err
	}

	tfID := vars.GetString("tf_id")
	if err := vars.Error(); err != nil {
		return "", err
	}

	newWorkspace, err := workspace.NewWorkspace(
		varsMap,
		"",
		action.Templates,
		evaluateParameterMappings(action.ImportParameterMappings),
		action.ImportParametersToDelete,
		evaluateParametersToAdd(action.ImportParametersToAdd))
	if err != nil {
		return tfID, fmt.Errorf("error creating workspace: %w", err)
	}

	deployment, err := provider.CreateAndSaveDeployment(tfID, newWorkspace)
	if err != nil {
		provider.logger.Error("deployment create failed", err)
		return tfID, fmt.Errorf("deployment create failed: %w", err)
	}

	if err := provider.MarkOperationStarted(&deployment, models.ProvisionOperationType); err != nil {
		return tfID, fmt.Errorf("error marking job started: %w", err)
	}

	go func() {
		logger := utils.NewLogger("Import").WithData(correlation.ID(ctx))
		resources := make(map[string]string)
		for _, resource := range importParams {
			resources[resource.TfResource] = resource.IaaSResource
		}

		terraformInvoker := provider.DefaultInvoker()
		var mainTf string
		err := steps.RunSequentially(
			func() (errs error) {
				return terraformInvoker.Import(ctx, newWorkspace, resources)
			},
			func() (errs error) {
				mainTf, err = terraformInvoker.Show(ctx, newWorkspace)
				return err
			},
			func() (errs error) {
				return createTFMainDefinition(newWorkspace, mainTf, logger)
			},
			func() (errs error) {
				return provider.terraformPlanToCheckNoResourcesDeleted(terraformInvoker, ctx, newWorkspace, logger)
			},
			func() (errs error) {
				if err := terraformInvoker.Apply(ctx, newWorkspace); err != nil {
					return err
				}
				_ = provider.MarkOperationFinished(&deployment, nil)
				return nil
			},
		)

		if err != nil {
			logger.Error("operation failed", err)
			_ = provider.MarkOperationFinished(&deployment, err)
		}
	}()

	return tfID, nil
}

func validateImportParams(importVariables []broker.ImportVariable, varsMap map[string]any) ([]ImportResource, error) {
	var importParams []ImportResource
	for _, importParam := range importVariables {
		if param, ok := varsMap[importParam.Name]; ok {
			importParams = append(importParams, ImportResource{TfResource: importParam.TfResource, IaaSResource: fmt.Sprintf("%v", param)})
		}
	}

	if len(importParams) != len(importVariables) {
		importFields := importVariables[0].Name
		for i := 1; i < len(importVariables); i++ {
			importFields = fmt.Sprintf("%s, %s", importFields, importVariables[i].Name)
		}

		return nil, fmt.Errorf("must provide values for all import parameters: %s", importFields)
	}

	return importParams, nil
}

func evaluateParameterMappings(importParameterMappings []ImportParameterMapping) []workspace.ParameterMapping {
	var parameterMappings []workspace.ParameterMapping
	for _, importParameterMapping := range importParameterMappings {
		parameterMappings = append(parameterMappings, workspace.ParameterMapping{
			TfVariable:    importParameterMapping.TfVariable,
			ParameterName: importParameterMapping.ParameterName,
		})
	}
	return parameterMappings
}

func evaluateParametersToAdd(importParametersToAdd []ImportParameterMapping) []workspace.ParameterMapping {
	var addParams []workspace.ParameterMapping
	for _, addParam := range importParametersToAdd {
		addParams = append(addParams, workspace.ParameterMapping{
			TfVariable:    addParam.TfVariable,
			ParameterName: addParam.ParameterName,
		})
	}
	return addParams
}

func createTFMainDefinition(workspace *workspace.TerraformWorkspace, mainTf string, logger lager.Logger) error {
	if i := strings.Index(mainTf, "\nOutputs:"); i >= 0 {
		mainTf = mainTf[:i]
	}

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
