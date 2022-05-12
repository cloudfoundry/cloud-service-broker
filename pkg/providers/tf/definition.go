// Copyright 2018 the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tf

import (
	"fmt"
	"os"
	"path"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/invoker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/validation"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/spf13/viper"
)

// NewExampleTfServiceDefinition creates a new service defintition with sample
// values for the service broker suitable to give a user a template to manually
// edit.
func NewExampleTfServiceDefinition() TfServiceDefinitionV1 {
	return TfServiceDefinitionV1{
		Version:          1,
		Name:             "example-service",
		ID:               "00000000-0000-0000-0000-000000000000",
		Description:      "a longer service description",
		DisplayName:      "Example Service",
		ImageURL:         "https://example.com/icon.jpg",
		DocumentationURL: "https://example.com",
		SupportURL:       "https://example.com/support.html",
		Tags:             []string{"gcp", "example", "service"},
		Plans: []TfServiceDefinitionV1Plan{
			{
				ID:          "00000000-0000-0000-0000-000000000001",
				Name:        "example-email-plan",
				DisplayName: "example.com email builder",
				Description: "Builds emails for example.com.",
				Bullets:     []string{"information point 1", "information point 2", "some caveat here"},
				Free:        false,
				Properties: map[string]interface{}{
					"domain":                 "example.com",
					"password_special_chars": `@/ \"?`,
				},
			},
		},
		ProvisionSettings: TfServiceDefinitionV1Action{
			PlanInputs: []broker.BrokerVariable{
				{
					FieldName: "domain",
					Type:      broker.JSONTypeString,
					Details:   "The domain name",
					Required:  true,
				},
			},
			UserInputs: []broker.BrokerVariable{
				{
					FieldName: "username",
					Type:      broker.JSONTypeString,
					Details:   "The username to create",
					Required:  true,
				},
			},
			Template: `
            variable domain {type = string}
            variable username {type = string}
            output email {value = "${var.username}@${var.domain}"}
			`,
			Outputs: []broker.BrokerVariable{
				{
					FieldName: "email",
					Type:      broker.JSONTypeString,
					Details:   "The combined email address",
					Required:  true,
				},
			},
		},
		BindSettings: TfServiceDefinitionV1Action{
			PlanInputs: []broker.BrokerVariable{
				{
					FieldName: "password_special_chars",
					Type:      broker.JSONTypeString,
					Details:   "Supply your own list of special characters to use for string generation.",
					Required:  true,
				},
			},
			Computed: []varcontext.DefaultVariable{
				{Name: "domain", Default: `${request.plan_properties["domain"]}`, Overwrite: true},
				{Name: "address", Default: `${instance.details["email"]}`, Overwrite: true},
			},
			Template: `
            variable domain {type = string}
            variable address {type = string}
            variable password_special_chars {type = string}

            resource "random_string" "password" {
                length = 16
                special = true
                override_special = var.password_special_chars
            }

            output uri {value = "smtp://${var.address}:${random_string.password.result}@smtp.${var.domain}"}
			`,
			Outputs: []broker.BrokerVariable{
				{
					FieldName: "uri",
					Type:      broker.JSONTypeString,
					Details:   "The uri to use to connect to this service",
					Required:  true,
				},
			},
		},
		Examples: []broker.ServiceExample{
			{
				Name:            "Example",
				Description:     "Examples are used for documenting your service AND as integration tests.",
				PlanID:          "00000000-0000-0000-0000-000000000001",
				ProvisionParams: map[string]interface{}{"username": "my-account"},
				BindParams:      map[string]interface{}{},
			},
		},
	}
}

// TfServiceDefinitionV1 is the first version of user defined services.
type TfServiceDefinitionV1 struct {
	Version           int                         `yaml:"version"`
	Name              string                      `yaml:"name"`
	ID                string                      `yaml:"id"`
	Description       string                      `yaml:"description"`
	DisplayName       string                      `yaml:"display_name"`
	ImageURL          string                      `yaml:"image_url"`
	DocumentationURL  string                      `yaml:"documentation_url"`
	SupportURL        string                      `yaml:"support_url"`
	Tags              []string                    `yaml:"tags,flow"`
	Plans             []TfServiceDefinitionV1Plan `yaml:"plans"`
	ProvisionSettings TfServiceDefinitionV1Action `yaml:"provision"`
	BindSettings      TfServiceDefinitionV1Action `yaml:"bind"`
	Examples          []broker.ServiceExample     `yaml:"examples"`
	PlanUpdateable    bool                        `yaml:"plan_updateable"`

	RequiredEnvVars []string
}

var _ validation.Validatable = (*TfServiceDefinitionV1)(nil)

// Validate checks the service definition for semantic errors.
func (tfb *TfServiceDefinitionV1) Validate() (errs *validation.FieldError) {

	if tfb.Version != 1 {
		errs = errs.Also(validation.ErrInvalidValue(tfb.Version, "version"))
	}

	errs = errs.Also(
		validation.ErrIfBlank(tfb.Name, "name"),
		validation.ErrIfNotUUID(tfb.ID, "id"),
		validation.ErrIfBlank(tfb.Description, "description"),
		validation.ErrIfBlank(tfb.DisplayName, "display_name"),
		validation.ErrIfNotURL(tfb.ImageURL, "image_url"),
		validation.ErrIfNotURL(tfb.DocumentationURL, "documentation_url"),
		validation.ErrIfNotURL(tfb.SupportURL, "support_url"),
	)

	names := make(map[string]struct{})
	ids := make(map[string]struct{})
	for i, v := range tfb.Plans {
		errs = errs.Also(
			v.Validate().ViaFieldIndex("plans", i),
			validation.ErrIfDuplicate(v.Name, "Name", names).ViaFieldIndex("plans", i),
			validation.ErrIfDuplicate(v.ID, "ID", ids).ViaFieldIndex("plans", i),
		)
	}

	errs = errs.Also(tfb.ProvisionSettings.Validate().ViaField("provision"))
	errs = errs.Also(tfb.BindSettings.Validate().ViaField("bind"))

	for i, v := range tfb.Examples {
		errs = errs.Also(v.Validate().ViaFieldIndex("examples", i))
	}

	return errs
}

func (tfb *TfServiceDefinitionV1) resolveEnvVars() (map[string]string, error) {
	vars := make(map[string]string)
	for _, v := range tfb.RequiredEnvVars {
		viper.BindEnv(v, v)
		if !viper.IsSet(v) {
			return vars, fmt.Errorf(fmt.Sprintf("missing required env var %s", v))
		}
		vars[v] = viper.GetString(v)
	}
	return vars, nil
}

func (tfb *TfServiceDefinitionV1) loadTemplates() error {
	err := tfb.BindSettings.LoadTemplate(".")

	if err == nil {
		err = tfb.ProvisionSettings.LoadTemplate(".")
	}

	return err
}

// ToService converts the flat TfServiceDefinitionV1 into a broker.ServiceDefinition
// that the registry can use.
func (tfb *TfServiceDefinitionV1) ToService(tfBinContext executor.TFBinariesContext) (*broker.ServiceDefinition, error) {
	if err := tfb.loadTemplates(); err != nil {
		return nil, err
	}

	if err := tfb.Validate(); err != nil {
		return nil, err
	}

	envVars, err := tfb.resolveEnvVars()
	if err != nil {
		return nil, err
	}

	var rawPlans []broker.ServicePlan
	for _, plan := range tfb.Plans {
		rawPlans = append(rawPlans, plan.ToPlan())
	}

	// Bindings get special computed properties because the broker didn't
	// originally support injecting plan variables into a binding
	// to fix that, we auto-inject the properties from the plan to make it look
	// like they were to the TF template.
	bindComputed := []varcontext.DefaultVariable{}
	for _, pi := range tfb.BindSettings.PlanInputs {
		bindComputed = append(bindComputed, varcontext.DefaultVariable{
			Name:      pi.FieldName,
			Default:   fmt.Sprintf("${request.plan_properties[%q]}", pi.FieldName),
			Overwrite: true,
			Type:      string(pi.Type),
		})
	}

	bindComputed = append(bindComputed, tfb.BindSettings.Computed...)
	bindComputed = append(bindComputed, varcontext.DefaultVariable{
		Name:      "tf_id",
		Default:   "tf:${request.instance_id}:${request.binding_id}",
		Overwrite: true,
	})

	constDefn := *tfb
	return &broker.ServiceDefinition{
		ID:               tfb.ID,
		Name:             tfb.Name,
		Description:      tfb.Description,
		Bindable:         true,
		PlanUpdateable:   tfb.PlanUpdateable,
		DisplayName:      tfb.DisplayName,
		DocumentationURL: tfb.DocumentationURL,
		SupportURL:       tfb.SupportURL,
		ImageURL:         tfb.ImageURL,
		Tags:             tfb.Tags,
		Plans:            rawPlans,

		ProvisionInputVariables: tfb.ProvisionSettings.UserInputs,
		ImportInputVariables:    tfb.ProvisionSettings.ImportVariables,
		ProvisionComputedVariables: append(tfb.ProvisionSettings.Computed, varcontext.DefaultVariable{
			Name:      "tf_id",
			Default:   "tf:${request.instance_id}:",
			Overwrite: true,
		}),
		BindInputVariables:    tfb.BindSettings.UserInputs,
		BindComputedVariables: bindComputed,
		BindOutputVariables:   append(tfb.ProvisionSettings.Outputs, tfb.BindSettings.Outputs...),
		PlanVariables:         append(tfb.ProvisionSettings.PlanInputs, tfb.BindSettings.PlanInputs...),
		Examples:              tfb.Examples,
		ProviderBuilder: func(logger lager.Logger, store broker.ServiceProviderStorage) broker.ServiceProvider {
			executorFactory := executor.NewExecutorFactory(tfBinContext.Dir, tfBinContext.Params, envVars)
			return NewTerraformProvider(NewTfJobRunner(store, tfBinContext, workspace.NewWorkspaceFactory(), invoker.NewTerraformInvokerFactory(executorFactory, tfBinContext.Dir, tfBinContext.ProviderReplacements)), logger, constDefn, store)
		},
	}, nil
}

// TfServiceDefinitionV1Plan represents a service plan in a human-friendly format
// that can be converted into an OSB compatible plan.
type TfServiceDefinitionV1Plan struct {
	Name               string                 `yaml:"name"`
	ID                 string                 `yaml:"id"`
	Description        string                 `yaml:"description"`
	DisplayName        string                 `yaml:"display_name"`
	Bullets            []string               `yaml:"bullets,omitempty"`
	Free               bool                   `yaml:"free,omitempty"`
	Properties         map[string]interface{} `yaml:"properties"`
	ProvisionOverrides map[string]interface{} `yaml:"provision_overrides,omitempty"`
	BindOverrides      map[string]interface{} `yaml:"bind_overrides,omitempty"`
}

var _ validation.Validatable = (*TfServiceDefinitionV1Plan)(nil)

// Validate implements validation.Validatable.
func (plan *TfServiceDefinitionV1Plan) Validate() (errs *validation.FieldError) {
	return errs.Also(
		validation.ErrIfBlank(plan.Name, "name"),
		validation.ErrIfNotUUID(plan.ID, "id"),
		validation.ErrIfBlank(plan.Description, "description"),
		validation.ErrIfBlank(plan.DisplayName, "display_name"),
	)
}

// ToPlan converts this plan definition to a broker.ServicePlan.
func (plan *TfServiceDefinitionV1Plan) ToPlan() broker.ServicePlan {
	masterPlan := domain.ServicePlan{
		ID:          plan.ID,
		Description: plan.Description,
		Name:        plan.Name,
		Free:        domain.FreeValue(plan.Free),
		Metadata: &domain.ServicePlanMetadata{
			Bullets:     plan.Bullets,
			DisplayName: plan.DisplayName,
		},
	}

	return broker.ServicePlan{
		ServicePlan:        masterPlan,
		ServiceProperties:  plan.Properties,
		ProvisionOverrides: plan.ProvisionOverrides,
		BindOverrides:      plan.BindOverrides,
	}
}

// TfServiceDefinitionV1Action holds information needed to process user inputs
// for a single provision or bind call.
type TfServiceDefinitionV1Action struct {
	PlanInputs               []broker.BrokerVariable      `yaml:"plan_inputs"`
	UserInputs               []broker.BrokerVariable      `yaml:"user_inputs"`
	Computed                 []varcontext.DefaultVariable `yaml:"computed_inputs"`
	Template                 string                       `yaml:"template"`
	TemplateRef              string                       `yaml:"template_ref"`
	Outputs                  []broker.BrokerVariable      `yaml:"outputs"`
	Templates                map[string]string            `yaml:"templates"`
	TemplateRefs             map[string]string            `yaml:"template_refs"`
	ImportVariables          []broker.ImportVariable      `yaml:"import_inputs"`
	ImportParameterMappings  []ImportParameterMapping     `yaml:"import_parameter_mappings"`
	ImportParametersToDelete []string                     `yaml:"import_parameters_to_delete"`
	ImportParametersToAdd    []ImportParameterMapping     `yaml:"import_parameters_to_add"`
}

var _ validation.Validatable = (*TfServiceDefinitionV1Action)(nil)

func (action *TfServiceDefinitionV1Action) IsTfImport(provisionContext *varcontext.VarContext) bool {
	const subsume = "subsume"
	for _, planInput := range action.PlanInputs {
		if planInput.FieldName == subsume && provisionContext.HasKey(subsume) && provisionContext.GetBool(subsume) {
			return true
		}
	}
	return false
}

func loadTemplate(templatePath string) (string, error) {
	if templatePath == "" {
		return "", nil
	}
	logger := lager.NewLogger("definition")
	logger.Info("loading template from:", lager.Data{"templatePath": templatePath})
	buff, err := os.ReadFile(templatePath)

	if err != nil {
		return "", err
	}
	return string(buff), nil
}

// LoadTemplate loads template ref into template if provided
func (action *TfServiceDefinitionV1Action) LoadTemplate(srcDir string) error {
	var err error

	if action.TemplateRef != "" {
		action.Template, err = loadTemplate(path.Join(srcDir, action.TemplateRef))
		if err != nil {
			return err
		}
	}

	if action.Templates == nil {
		action.Templates = make(map[string]string)
	}

	for name, ref := range action.TemplateRefs {
		if ref != "" {
			action.Templates[name], err = loadTemplate(path.Join(srcDir, ref))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Validate implements validation.Validatable.
func (action *TfServiceDefinitionV1Action) Validate() (errs *validation.FieldError) {
	for i, v := range action.PlanInputs {
		errs = errs.Also(v.Validate().ViaFieldIndex("plan_inputs", i))
	}

	for i, v := range action.UserInputs {
		errs = errs.Also(v.Validate().ViaFieldIndex("user_inputs", i))
	}

	for i, v := range action.Computed {
		errs = errs.Also(v.Validate().ViaFieldIndex("computed_inputs", i))
	}

	if action.TemplateRef != "" {
		errs = errs.Also(validation.ErrIfBlank(action.Template, "template not loaded from templat ref"))
	}

	errs = errs.Also(
		validation.ErrIfNotHCL(action.Template, "template"),
		action.validateTemplateInputs().ViaField("template"),
		action.validateTemplateOutputs().ViaField("template"),
	)

	for i, v := range action.Outputs {
		errs = errs.Also(v.Validate().ViaFieldIndex("outputs", i))
	}

	return errs
}

func (action *TfServiceDefinitionV1Action) ValidateTemplateIO() (errs *validation.FieldError) {
	return errs.Also(
		action.validateTemplateInputs().ViaField("template"),
		action.validateTemplateOutputs().ViaField("template"),
	)
}

// validateTemplateInputs checks that all the inputs of the Terraform template
// are defined by the service.
func (action *TfServiceDefinitionV1Action) validateTemplateInputs() (errs *validation.FieldError) {
	inputs := utils.NewStringSet()

	for _, in := range action.PlanInputs {
		inputs.Add(in.FieldName)
	}

	for _, in := range action.UserInputs {
		inputs.Add(in.FieldName)
	}

	for _, in := range action.Computed {
		inputs.Add(in.Name)
	}

	tfModule := workspace.ModuleDefinition{Definition: action.Template, Definitions: action.Templates}
	tfIn, err := tfModule.Inputs()
	if err != nil {
		return &validation.FieldError{
			Message: err.Error(),
		}
	}

	missingFields := utils.NewStringSet(tfIn...).Minus(inputs).ToSlice()
	if len(missingFields) > 0 {
		return &validation.FieldError{
			Message: "fields used but not declared",
			Paths:   missingFields,
		}
	}

	return nil
}

// validateTemplateOutputs checks that the Terraform template outputs match
// the names of the defined outputs.
func (action *TfServiceDefinitionV1Action) validateTemplateOutputs() (errs *validation.FieldError) {
	definedOutputs := utils.NewStringSet()

	for _, in := range action.Outputs {
		definedOutputs.Add(in.FieldName)
	}

	tfModule := workspace.ModuleDefinition{Definition: action.Template, Definitions: action.Templates}
	tfOut, err := tfModule.Outputs()
	if err != nil {
		return &validation.FieldError{
			Message: err.Error(),
		}
	}

	if !definedOutputs.Equals(utils.NewStringSet(tfOut...).Minus(utils.NewStringSet("status"))) {
		return &validation.FieldError{
			Message: fmt.Sprintf("template outputs %v must match declared outputs %v", tfOut, definedOutputs),
		}
	}

	return nil
}

// generateTfID creates a unique id for a given provision/bind combination that
// will be consistent across calls. This ID will be used in LastOperation polls
// as well as to uniquely identify the workspace.
func generateTfID(instanceID, bindingID string) string {
	return fmt.Sprintf("tf:%s:%s", instanceID, bindingID)
}

// ImportParameterMapping mapping for tf variable to service parameter
type ImportParameterMapping struct {
	TfVariable    string `yaml:"tf_variable"`
	ParameterName string `yaml:"parameter_name"`
}

type TfCatalogDefinitionV1 []*TfServiceDefinitionV1

var _ validation.Validatable = (*TfCatalogDefinitionV1)(nil)

// Validate checks the service definitions for semantic errors.
func (tfb TfCatalogDefinitionV1) Validate() (errs *validation.FieldError) {
	names := make(map[string]struct{})
	serviceIDs := make(map[string]struct{})
	planIDs := make(map[string]struct{})

	for i, service := range tfb {
		errs = errs.Also(
			service.Validate().ViaFieldIndex("services", i),
			validation.ErrIfDuplicate(service.Name, "Name", names).ViaFieldIndex("services", i),
			validation.ErrIfDuplicate(service.ID, "ID", serviceIDs).ViaFieldIndex("services", i),
		)

		for j, plan := range service.Plans {
			errs = errs.Also(validation.ErrIfDuplicate(plan.ID, "ID", planIDs)).ViaFieldIndex("plans", j).ViaFieldIndex("services", i)
		}
	}

	return errs
}
