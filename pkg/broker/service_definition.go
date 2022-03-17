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

package broker

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/toggles"
	"github.com/cloudfoundry/cloud-service-broker/pkg/validation"
	"github.com/cloudfoundry/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/spf13/viper"
)

var enableCatalogSchemas = toggles.Features.Toggle("enable-catalog-schemas", false, `Enable generating JSONSchema for the service catalog.`)

// GlobalProvisionDefaults viper key for global provision defaults
const GlobalProvisionDefaults = "provision.defaults"

// ServiceDefinition holds the necessary details to describe an OSB service and
// provision it.
type ServiceDefinition struct {
	Id               string
	Name             string
	Description      string
	DisplayName      string
	ImageUrl         string
	DocumentationUrl string
	SupportUrl       string
	Tags             []string
	Bindable         bool
	PlanUpdateable   bool
	Plans            []ServicePlan

	ProvisionInputVariables    []BrokerVariable
	ImportInputVariables       []ImportVariable
	ProvisionComputedVariables []varcontext.DefaultVariable
	BindInputVariables         []BrokerVariable
	BindOutputVariables        []BrokerVariable
	BindComputedVariables      []varcontext.DefaultVariable
	PlanVariables              []BrokerVariable
	Examples                   []ServiceExample
	DefaultRoleWhitelist       []string

	// ProviderBuilder creates a new provider given the project, auth, and logger.
	ProviderBuilder func(plogger lager.Logger, store ServiceProviderStorage) ServiceProvider

	// IsBuiltin is true if the service is built-in to the platform.
	IsBuiltin bool
}

var _ validation.Validatable = (*ServiceDefinition)(nil)

// Validate implements validation.Validatable.
func (svc *ServiceDefinition) Validate() (errs *validation.FieldError) {
	errs = errs.Also(
		validation.ErrIfNotUUID(svc.Id, "Id"),
		validation.ErrIfNotOSBName(svc.Name, "Name"),
	)

	if svc.ImageUrl != "" {
		errs = errs.Also(validation.ErrIfNotURL(svc.ImageUrl, "ImageUrl"))
	}

	if svc.DocumentationUrl != "" {
		errs = errs.Also(validation.ErrIfNotURL(svc.DocumentationUrl, "DocumentationUrl"))
	}

	if svc.SupportUrl != "" {
		errs = errs.Also(validation.ErrIfNotURL(svc.SupportUrl, "SupportUrl"))
	}

	for i, v := range svc.ProvisionInputVariables {
		errs = errs.Also(v.Validate().ViaFieldIndex("ProvisionInputVariables", i))
	}

	for i, v := range svc.ProvisionComputedVariables {
		errs = errs.Also(v.Validate().ViaFieldIndex("ProvisionComputedVariables", i))
	}

	for i, v := range svc.BindInputVariables {
		errs = errs.Also(v.Validate().ViaFieldIndex("BindInputVariables", i))
	}

	for i, v := range svc.BindOutputVariables {
		errs = errs.Also(v.Validate().ViaFieldIndex("BindOutputVariables", i))
	}

	for i, v := range svc.BindComputedVariables {
		errs = errs.Also(v.Validate().ViaFieldIndex("BindComputedVariables", i))
	}

	for i, v := range svc.PlanVariables {
		errs = errs.Also(v.Validate().ViaFieldIndex("PlanVariables", i))
	}

	names := make(map[string]struct{})
	ids := make(map[string]struct{})
	for i, v := range svc.Plans {
		errs = errs.Also(
			v.Validate().ViaFieldIndex("Plans", i),
			validation.ErrIfDuplicate(v.Name, "Name", names).ViaFieldIndex("Plans", i),
			validation.ErrIfDuplicate(v.ID, "Id", ids).ViaFieldIndex("Plans", i),
		)
	}

	return errs
}

// UserDefinedPlansProperty computes the Viper property name for the JSON list
// of user-defined service plans.
func (svc *ServiceDefinition) UserDefinedPlansProperty() string {
	return fmt.Sprintf("service.%s.plans", svc.Name)
}

func (svc *ServiceDefinition) UserDefinedPlansVariable() string {
	return strings.ToUpper(
		strings.ReplaceAll(
			fmt.Sprintf("%s.service.%s.plans", utils.EnvironmentVarPrefix, utils.PropertyToEnvUnprefixed(svc.Name)),
			".",
			"_",
		),
	)
}

// ProvisionDefaultOverrideProperty returns the Viper property name for the
// object users can set to override the default values on provision.
func (svc *ServiceDefinition) ProvisionDefaultOverrideProperty() string {
	return fmt.Sprintf("service.%s.provision.defaults", svc.Name)
}

// ProvisionDefaultOverrides returns the deserialized JSON object for the
// operator-provided property overrides.
func (svc *ServiceDefinition) ProvisionDefaultOverrides() (map[string]interface{}, error) {
	return unmarshalViper(svc.ProvisionDefaultOverrideProperty())
}

func ProvisionGlobalDefaults() (map[string]interface{}, error) {
	return unmarshalViper(GlobalProvisionDefaults)
}

func unmarshalViper(key string) (map[string]interface{}, error) {
	vals := make(map[string]interface{})
	if viper.IsSet(key) {
		val := viper.GetString(key)
		if err := json.Unmarshal([]byte(val), &vals); err != nil {
			return nil, fmt.Errorf("failed unmarshaling config value %s", key)
		}
	}
	return vals, nil
}

// IsRoleWhitelistEnabled returns false if the service has no default whitelist
// meaning it does not allow any roles.
func (svc *ServiceDefinition) IsRoleWhitelistEnabled() bool {
	return len(svc.DefaultRoleWhitelist) > 0
}

// BindDefaultOverrideProperty returns the Viper property name for the
// object users can set to override the default values on bind.
func (svc *ServiceDefinition) BindDefaultOverrideProperty() string {
	return fmt.Sprintf("service.%s.bind.defaults", svc.Name)
}

// BindDefaultOverrides returns the deserialized JSON object for the
// operator-provided property overrides.
func (svc *ServiceDefinition) BindDefaultOverrides() map[string]interface{} {
	return viper.GetStringMap(svc.BindDefaultOverrideProperty())
}

// TileUserDefinedPlansVariable returns the name of the user defined plans
// variable for the broker tile.
func (svc *ServiceDefinition) TileUserDefinedPlansVariable() string {
	v := utils.PropertyToEnvUnprefixed(svc.Name)
	v = strings.TrimPrefix(v, "GOOGLE_")

	return v + "_CUSTOM_PLANS"
}

// CatalogEntry returns the service broker catalog entry for this service, it
// has metadata about the service so operators and programmers know which
// service and plan will work best for their purposes.
func (svc *ServiceDefinition) CatalogEntry() *Service {
	sd := &Service{
		Service: domain.Service{
			ID:          svc.Id,
			Name:        svc.Name,
			Description: svc.Description,
			Metadata: &domain.ServiceMetadata{
				DisplayName:     svc.DisplayName,
				LongDescription: svc.Description,

				DocumentationUrl: svc.DocumentationUrl,
				ImageUrl:         svc.ImageUrl,
				SupportUrl:       svc.SupportUrl,
			},
			Tags:          svc.Tags,
			Bindable:      svc.Bindable,
			PlanUpdatable: svc.PlanUpdateable,
		},
		Plans: svc.Plans,
	}

	if enableCatalogSchemas.IsActive() {
		for i := range sd.Plans {
			sd.Plans[i].Schemas = svc.createSchemas()
		}
	}

	return sd
}

// createSchemas creates JSONSchemas compatible with the OSB spec for provision and bind.
// It leaves the instance update schema empty to indicate updates are not supported.
func (svc *ServiceDefinition) createSchemas() *domain.ServiceSchemas {
	return &domain.ServiceSchemas{
		Instance: domain.ServiceInstanceSchema{
			Create: domain.Schema{
				Parameters: CreateJsonSchema(svc.ProvisionInputVariables),
			},
		},
		Binding: domain.ServiceBindingSchema{
			Create: domain.Schema{
				Parameters: CreateJsonSchema(svc.BindInputVariables),
			},
		},
	}
}

// GetPlanById finds a plan in this service by its UUID.
func (svc *ServiceDefinition) GetPlanById(planId string) (*ServicePlan, error) {
	catalogEntry := svc.CatalogEntry()
	for _, plan := range catalogEntry.Plans {
		if plan.ID == planId {
			return &plan, nil
		}
	}

	return nil, fmt.Errorf("plan ID %q could not be found", planId)
}

// UserDefinedPlans extracts user defined plans from the environment, failing if
// the plans were not valid JSON or were missing required properties/variables.
func (svc *ServiceDefinition) UserDefinedPlans() ([]ServicePlan, error) {

	// There's a mismatch between how plans are used internally and defined by
	// the user and the tile. In the environment variables we parse an array of
	// flat maps, but internally extra variables need to be put into a sub-map.
	// e.g. they come in as [{"id":"1234", "name":"foo", "A": 1, "B": 2}]
	// but we need [{"id":"1234", "name":"foo", "service_properties":{"A": 1, "B": 2}}]
	// Go doesn't support this natively so we do it manually here.
	rawPlans := []json.RawMessage{}

	// Unmarshal the plans from the viper configuration which is just a JSON list
	// of plans
	if userPlanJSON := viper.GetString(svc.UserDefinedPlansProperty()); userPlanJSON != "" {
		if err := json.Unmarshal([]byte(userPlanJSON), &rawPlans); err != nil {
			return []ServicePlan{}, err
		}
	}

	// Unmarshal tile plans if they're included, which are a JSON object where
	// keys are
	if tilePlans := os.Getenv(svc.TileUserDefinedPlansVariable()); tilePlans != "" {
		var rawTilePlans map[string]json.RawMessage
		if err := json.Unmarshal([]byte(tilePlans), &rawTilePlans); err != nil {
			return []ServicePlan{}, err
		}

		for _, v := range rawTilePlans {
			rawPlans = append(rawPlans, v)
		}
	}

	plans := []ServicePlan{}
	for _, rawPlan := range rawPlans {
		plan := ServicePlan{}
		remainder, err := utils.UnmarshalObjectRemainder(rawPlan, &plan)
		if err != nil {
			return []ServicePlan{}, err
		}

		plan.ServiceProperties = make(map[string]interface{})
		if err := json.Unmarshal(remainder, &plan.ServiceProperties); err != nil {
			return []ServicePlan{}, err
		}

		// reading from a tile we need to move their GUID to an ID field
		if plan.ID == "" {
			plan.ID, _ = plan.ServiceProperties["guid"].(string)
		}

		if err := svc.validatePlan(plan); err != nil {
			return []ServicePlan{}, err
		}

		plans = append(plans, plan)
	}

	return plans, nil
}

func (svc *ServiceDefinition) validatePlan(plan ServicePlan) error {
	if plan.ID == "" {
		return fmt.Errorf("%s custom plan %+v is missing an id", svc.Name, plan)
	}

	if plan.Name == "" {
		return fmt.Errorf("%s custom plan %+v is missing a name", svc.Name, plan)
	}

	if svc.PlanVariables == nil {
		return nil
	}

	for _, customVar := range svc.PlanVariables {
		if !customVar.Required {
			continue
		}

		if _, ok := plan.ServiceProperties[customVar.FieldName]; !ok {
			return fmt.Errorf("%s custom plan %+v is missing required property %s", svc.Name, plan, customVar.FieldName)
		}
	}

	return nil
}

func (svc *ServiceDefinition) provisionDefaults() []varcontext.DefaultVariable {
	var out []varcontext.DefaultVariable
	for _, provisionVar := range svc.ProvisionInputVariables {
		out = append(out, varcontext.DefaultVariable{Name: provisionVar.FieldName, Default: provisionVar.Default, Overwrite: false, Type: string(provisionVar.Type)})
	}
	return out
}

func (svc *ServiceDefinition) bindDefaults() []varcontext.DefaultVariable {
	var out []varcontext.DefaultVariable
	for _, v := range svc.BindInputVariables {
		out = append(out, varcontext.DefaultVariable{Name: v.FieldName, Default: v.Default, Overwrite: false, Type: string(v.Type)})
	}
	return out
}

// ProvisionVariables gets the variable resolution context for a provision request.
// Variables have a very specific resolution order, and this function populates the context to preserve that.
// The variable resolution order is the following:
//
// 1. Variables defined in your `computed_variables` JSON list.
// 2. Variables defined by the selected service plan in its `service_properties` map.
// 3. Variables overridden in the plan's `provision_overrides` map.
// 4. User defined variables (in `update_input_variables`)
// 5. User defined variables (in `provision_input_variables` or `bind_input_variables`)
// 6. Operator default variables loaded from the environment.
// 7. Global operator default variables loaded from the environment.
// 8. Default variables (in `provision_input_variables` or `bind_input_variables`).
//
// Loading into the map occurs slightly differently.
// Default variables and computed_variables get executed by interpolation.
// User defined variables are not to prevent side-channel attacks.
// Default variables may reference user provided variables.
// For example, to create a default database name based on a user-provided instance name.
// Therefore, they get executed conditionally if a user-provided variable does not exist.
// Computed variables get executed either unconditionally or conditionally for greater flexibility.
func (svc *ServiceDefinition) variables(
	constants map[string]interface{},
	userProvidedParameters map[string]interface{},
	plan ServicePlan) (*varcontext.VarContext, error) {

	globalDefaults, err := ProvisionGlobalDefaults()
	if err != nil {
		return nil, err
	}
	provisionDefaultOverrides, err := svc.ProvisionDefaultOverrides()
	if err != nil {
		return nil, err
	}
	builder := varcontext.Builder().
		SetEvalConstants(constants).
		MergeMap(globalDefaults).
		MergeMap(provisionDefaultOverrides).
		MergeMap(userProvidedParameters).
		MergeMap(plan.ProvisionOverrides).
		MergeDefaults(svc.provisionDefaults()).
		MergeMap(plan.GetServiceProperties()).
		MergeDefaults(svc.ProvisionComputedVariables)

	return buildAndValidate(builder, svc.ProvisionInputVariables)
}

func (svc *ServiceDefinition) ProvisionVariables(instanceID string, details paramparser.ProvisionDetails, plan ServicePlan, originatingIdentity map[string]interface{}) (*varcontext.VarContext, error) {
	// The namespaces of these values roughly align with the OSB spec.
	constants := map[string]interface{}{
		"request.plan_id":     details.PlanID,
		"request.service_id":  details.ServiceID,
		"request.instance_id": instanceID,
		"request.default_labels": map[string]string{
			"pcf-organization-guid": utils.InvalidLabelChars.ReplaceAllString(details.OrganizationGUID, "_"),
			"pcf-space-guid":        utils.InvalidLabelChars.ReplaceAllString(details.SpaceGUID, "_"),
			"pcf-instance-id":       utils.InvalidLabelChars.ReplaceAllString(instanceID, "_"),
		},
		"request.context": details.RequestContext,
		"request.x_broker_api_originating_identity": originatingIdentity,
	}

	return svc.variables(constants, details.RequestParams, plan)
}

func (svc *ServiceDefinition) UpdateVariables(instanceId string, details paramparser.UpdateDetails, mergedUserProvidedParameters map[string]interface{}, plan ServicePlan, originatingIdentity map[string]interface{}) (*varcontext.VarContext, error) {
	constants := map[string]interface{}{
		"request.plan_id":     details.PlanID,
		"request.service_id":  details.ServiceID,
		"request.instance_id": instanceId,
		"request.default_labels": map[string]string{
			"pcf-organization-guid": utils.InvalidLabelChars.ReplaceAllString(details.PreviousOrgID, "_"),
			"pcf-space-guid":        utils.InvalidLabelChars.ReplaceAllString(details.PreviousSpaceID, "_"),
			"pcf-instance-id":       utils.InvalidLabelChars.ReplaceAllString(instanceId, "_"),
		},
		"request.context": details.RequestContext,
		"request.x_broker_api_originating_identity": originatingIdentity,
	}
	return svc.variables(constants, mergedUserProvidedParameters, plan)
}

// BindVariables gets the variable resolution context for a bind request.
// Variables have a very specific resolution order, and this function populates the context to preserve that.
// The variable resolution order is the following:
//
// 1. Variables defined in your `computed_variables` JSON list.
// 2. Variables overridden in the plan's `bind_overrides` map.
// 3. User defined variables (in `bind_input_variables`)
// 4. Operator default variables loaded from the environment.
// 5. Default variables (in `bind_input_variables`).
//
func (svc *ServiceDefinition) BindVariables(instance storage.ServiceInstanceDetails, bindingID string, details paramparser.BindDetails, plan *ServicePlan, originatingIdentity map[string]interface{}) (*varcontext.VarContext, error) {
	var appGuid string
	if details.BindAppGuid != "" {
		appGuid = details.BindAppGuid
	}

	// The namespaces of these values roughly align with the OSB spec.
	constants := map[string]interface{}{
		"request.x_broker_api_originating_identity": originatingIdentity,

		// specified in the URL
		"request.binding_id":  bindingID,
		"request.instance_id": instance.GUID,
		"request.context":     details.RequestContext,

		// specified in the request body
		// Note: the value in instance is considered the official record so values
		// are pulled from there rather than the request. In a future version of OSB
		// the duplicate sending of fields is likely to be removed.
		"request.plan_id":         instance.PlanGUID,
		"request.service_id":      instance.ServiceGUID,
		"request.app_guid":        appGuid,
		"request.plan_properties": plan.GetServiceProperties(),

		// specified by the existing instance
		"instance.name":    instance.Name,
		"instance.details": instance.Outputs,
	}

	builder := varcontext.Builder().
		SetEvalConstants(constants).
		MergeMap(svc.BindDefaultOverrides()).
		MergeMap(details.RequestParams).
		MergeMap(plan.BindOverrides).
		MergeDefaults(svc.bindDefaults()).
		MergeDefaults(svc.BindComputedVariables)

	return buildAndValidate(builder, svc.BindInputVariables)
}

// buildAndValidate builds the varcontext and if it's valid validates the
// resulting context against the JSONSchema defined by the BrokerVariables
// exactly one of VarContext and error will be nil upon return.
func buildAndValidate(builder *varcontext.ContextBuilder, vars []BrokerVariable) (*varcontext.VarContext, error) {
	vc, err := builder.Build()
	if err != nil {
		return nil, err
	}

	if err := ValidateVariables(vc.ToMap(), vars); err != nil {
		return nil, err
	}

	return vc, nil
}

func (svc *ServiceDefinition) AllowedUpdate(params map[string]interface{}) (bool, error) {
	for _, param := range svc.ProvisionInputVariables {
		if param.ProhibitUpdate {
			if _, ok := params[param.FieldName]; ok {
				return false, nil
			}
		}
	}
	return true, nil
}
