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
    "reflect"
    "os"
    "fmt"
    "strings"
    "testing"

    "github.com/go-yaml/yaml"
    "github.com/pivotal/cloud-service-broker/pkg/broker"
    "github.com/pivotal/cloud-service-broker/pkg/varcontext"
    "github.com/pivotal-cf/brokerapi"
)

func TestUnmarshalDefinition(t *testing.T) {
    cases := map[string]struct {
        yaml string
        expected TfServiceDefinitionV1
    }{
        "template": {
            yaml: `version: 1
name: my-service
id: fa22af0f-3637-4a36-b8a7-cfc61168a3e0
description: Some service
display_name: Service for some resource
image_url: image-urldt
documentation_url: documentation-url
support_url: support-url
tags: [iaas, service]
plans:
- name: small
  id: 2268ce43-7fd7-48dc-be2f-8611e11fb12e
  description: 'A small my-service'
  display_name: "small"
  properties:
    plan_prop_1: "value1"
provision:
  plan_inputs:
  - field_name: plan_prop_1
    required: true
    type: string
    details: Plan property 1
  user_inputs:
  - field_name: user_prop_1
    type: number
    details: User property 1
  computed_inputs:
  - name: computed_prop_1
    default: computed_default
    type: string
  template: |-
    variable plan_prop_1 { type = string }
    variable user_prop_1 { type = number }
    variable computed_prop_1 { type = string }
    output output_prop_1 { value = var.plan_prop_1 }
  outputs:
  - field_name: output_prop_1
    type: string
    details: Plan property 1
bind:
  plan_inputs: []
  user_inputs: []
  computed_inputs: []
  outputs: []
examples:
- name: small
  description: A small service instance
  plan_id: 2268ce43-7fd7-48dc-be2f-8611e11fb12e
  provision_params: {}
  bind_params: {}
    `,
            expected: TfServiceDefinitionV1{
                Version: 1,
                Name: "my-service",
                Id: "fa22af0f-3637-4a36-b8a7-cfc61168a3e0",
                Description: "Some service",
                DisplayName: "Service for some resource",
                ImageUrl: "image-url",
                DocumentationUrl: "documentation-url",
                SupportUrl: "support-url",
                Tags: []string{ "iaas", "service"},
                Plans: []TfServiceDefinitionV1Plan{
                    {
                        Name: "small",
                        Id: "2268ce43-7fd7-48dc-be2f-8611e11fb12e",
                        Description: "A small my-service",
                        DisplayName: "small",
                        Properties: map[string]interface{} {
                            "plan_prop_1": "value1",
                        },
                    },
                },
                ProvisionSettings: TfServiceDefinitionV1Action{
                    PlanInputs: []broker.BrokerVariable{
                        {
                            FieldName: "plan_prop_1",
                            Required: true,
                            Type: broker.JsonTypeString,
                            Details: "Plan property 1",
                        },
                    },
                    UserInputs: []broker.BrokerVariable{
                        {
                            FieldName: "user_prop_1",
                            Type: broker.JsonTypeNumeric,
                            Details: "User property 1",
                        },
                    },
                    Computed: []varcontext.DefaultVariable{
                        {
                            Name: "computed_prop_1",
                            Default: "computed_default",
                            Type: "string",
                        },
                    },
                    Template: `variable plan_prop_1 { type = string }
variable user_prop_1 { type = number }
variable computed_prop_1 { type = string }
output output_prop_1 { value = var.plan_prop_1 }`,
                    Outputs: []broker.BrokerVariable{
                        {
                            FieldName: "output_prop_1",
                            Type: broker.JsonTypeString,
                            Details: "Plan property 1",
                        },
                    },
                },
                BindSettings: TfServiceDefinitionV1Action{

                },
                Examples: []broker.ServiceExample{
                    {
                        Name: "small",
                        Description: "A small service instance",
                        PlanId: "2268ce43-7fd7-48dc-be2f-8611e11fb12e",
                    },
                } ,
            },
        },
        "templates": {
            yaml: `version: 1
name: my-service
id: fa22af0f-3637-4a36-b8a7-cfc61168a3e0
description: Some service
display_name: Service for some resource
image_url: image-urldt
documentation_url: documentation-url
support_url: support-url
tags: [iaas, service]
plans:
- name: small
  id: 2268ce43-7fd7-48dc-be2f-8611e11fb12e
  description: 'A small my-service'
  display_name: "small"
  properties:
    plan_prop_1: "value1"
provision:
  plan_inputs:
  - field_name: plan_prop_1
    required: true
    type: string
    details: Plan property 1
  user_inputs:
  - field_name: user_prop_1
    type: number
    details: User property 1
  computed_inputs:
  - name: computed_prop_1
    default: computed_default
    type: string
  outputs:
  - field_name: output_prop_1
    type: string
    details: Plan property 1
  templates:
    variables: "variable plan_prop_1 { type = string }"
    outputs: "output output_prop_1 { value = var.plan_prop_1 }"
  outputs_tf_ref: outputs_ref
  provider_tf_ref: provider_ref
  variables_tf_ref: variables_ref
  main_tf_ref: main_ref
bind:
  plan_inputs: []
  user_inputs: []
  computed_inputs: []
  outputs: []
examples:
- name: small
  description: A small service instance
  plan_id: 2268ce43-7fd7-48dc-be2f-8611e11fb12e
  provision_params: {}
  bind_params: {}
    `,
            expected: TfServiceDefinitionV1{
                Version: 1,
                Name: "my-service",
                Id: "fa22af0f-3637-4a36-b8a7-cfc61168a3e0",
                Description: "Some service",
                DisplayName: "Service for some resource",
                ImageUrl: "image-url",
                DocumentationUrl: "documentation-url",
                SupportUrl: "support-url",
                Tags: []string{ "iaas", "service"},
                Plans: []TfServiceDefinitionV1Plan{
                    {
                        Name: "small",
                        Id: "2268ce43-7fd7-48dc-be2f-8611e11fb12e",
                        Description: "A small my-service",
                        DisplayName: "small",
                        Properties: map[string]interface{} {
                            "plan_prop_1": "value1",
                        },
                    },
                },
                ProvisionSettings: TfServiceDefinitionV1Action{
                    PlanInputs: []broker.BrokerVariable{
                        {
                            FieldName: "plan_prop_1",
                            Required: true,
                            Type: broker.JsonTypeString,
                            Details: "Plan property 1",
                        },
                    },
                    UserInputs: []broker.BrokerVariable{
                        {
                            FieldName: "user_prop_1",
                            Type: broker.JsonTypeNumeric,
                            Details: "User property 1",
                        },
                    },
                    Computed: []varcontext.DefaultVariable{
                        {
                            Name: "computed_prop_1",
                            Default: "computed_default",
                            Type: "string",
                        },
                    },
                    Outputs: []broker.BrokerVariable{
                        {
                            FieldName: "output_prop_1",
                            Type: broker.JsonTypeString,
                            Details: "Plan property 1",
                        },
                    },
                    Templates: map[string]string {
                        "variables": "variable plan_prop_1 { type = string }",
                        "outputs": "output output_prop_1 { value = var.plan_prop_1 }",
                    },
                    OutputsRef: "outputs_ref",
                    ProviderRef: "provider_ref",
                    VariablesRef: "variables_ref",
                    MainRef: "main_ref",
                },
                BindSettings: TfServiceDefinitionV1Action{

                },
                Examples: []broker.ServiceExample{
                    {
                        Name: "small",
                        Description: "A small service instance",
                        PlanId: "2268ce43-7fd7-48dc-be2f-8611e11fb12e",
                    },
                } ,
            },
        },        
    }

    for tn, tc := range cases {
        t.Run(tn, func(t *testing.T) {
            var actual TfServiceDefinitionV1
            err := yaml.Unmarshal([]byte(tc.yaml), &actual)
            if err != nil {
                t.Fatalf("failed to unmarshal yaml definition: %v", err)
            }

            if !reflect.DeepEqual(actual.Tags, tc.expected.Tags) {
                t.Fatalf("Tags Expected: %+v Actual: %+v", tc.expected.Tags, actual.Tags)
            }
            if !reflect.DeepEqual(actual.Plans, tc.expected.Plans) {
                t.Fatalf("Plans Expected: %+v Actual: %+v", tc.expected.Plans, actual.Plans)
            }
            if !reflect.DeepEqual(actual.ProvisionSettings, tc.expected.ProvisionSettings) {
                t.Fatalf("ProvisionSettings Expected: %+v Actual: %+v", tc.expected.ProvisionSettings, actual.ProvisionSettings)
            }
        })
    }
}

func TestTfServiceDefinitionV1Action_ValidateTemplateIO(t *testing.T) {
    cases := map[string]struct {
        Action      TfServiceDefinitionV1Action
        ErrContains string
    }{
        "nomainal": {
            Action: TfServiceDefinitionV1Action{
                PlanInputs: []broker.BrokerVariable{{FieldName: "storage_class"}},
                UserInputs: []broker.BrokerVariable{{FieldName: "name"}},
                Computed:   []varcontext.DefaultVariable{{Name: "labels"}},
                Template: `
          variable storage_class {type = "string"}
          variable name {type = "string"}
          variable labels {type = "string"}

          output bucket_name {value = "${var.name}"}
          `,
                Outputs: []broker.BrokerVariable{{FieldName: "bucket_name"}},
            },
            ErrContains: "",
        },
        "extra inputs okay": {
            Action: TfServiceDefinitionV1Action{
                PlanInputs: []broker.BrokerVariable{{FieldName: "storage_class"}},
                UserInputs: []broker.BrokerVariable{{FieldName: "name"}},
                Computed:   []varcontext.DefaultVariable{{Name: "labels"}},
                Template: `
          variable storage_class {type = "string"}
          `,
            },
            ErrContains: "",
        },
        "missing inputs": {
            Action: TfServiceDefinitionV1Action{
                PlanInputs: []broker.BrokerVariable{{FieldName: "storage_class"}},
                UserInputs: []broker.BrokerVariable{{FieldName: "name"}},
                Computed:   []varcontext.DefaultVariable{{Name: "labels"}},
                Template: `
        variable storage_class {type = "string"}
        variable not_defined {type = "string"}
        `,
            },
            ErrContains: "fields used but not declared: template.not_defined",
        },

        "extra template outputs": {
            Action: TfServiceDefinitionV1Action{
                Template: `
        output storage_class {value = "${var.name}"}
        output name {value = "${var.name}"}
        output labels {value = "${var.name}"}
        output bucket_name {value = "${var.name}"}
        `,
                Outputs: []broker.BrokerVariable{{FieldName: "bucket_name"}},
            },
            ErrContains: "template outputs [bucket_name labels name storage_class] must match declared outputs [bucket_name]:",
        },

        "missing template outputs": {
            Action: TfServiceDefinitionV1Action{
                Template: `
        `,
                Outputs: []broker.BrokerVariable{{FieldName: "bucket_name"}},
            },
            ErrContains: "template outputs [] must match declared outputs [bucket_name]:",
        },
    }

    for tn, tc := range cases {
        t.Run(tn, func(t *testing.T) {
            err := tc.Action.ValidateTemplateIO()
            if err == nil {
                if tc.ErrContains == "" {
                    return
                }

                t.Fatalf("Expected error to contain %q, got: <nil>", tc.ErrContains)
            } else {
                if tc.ErrContains == "" {
                    t.Fatalf("Expected no error, got: %v", err)
                }

                if !strings.Contains(err.Error(), tc.ErrContains) {
                    t.Fatalf("Expected error to contain %q, got: %v", tc.ErrContains, err)
                }
            }
        })
    }
}

func TestTfServiceDefinitionV1Action_ValidateTemplatesIO(t *testing.T) {
    cases := map[string]struct {
        Action      TfServiceDefinitionV1Action
        ErrContains string
    }{
        "nomainal": {
            Action: TfServiceDefinitionV1Action{
                PlanInputs: []broker.BrokerVariable{{FieldName: "storage_class"}},
                UserInputs: []broker.BrokerVariable{{FieldName: "name"}},
                Computed:   []varcontext.DefaultVariable{{Name: "labels"}},
                Templates: map[string]string {
                    "variables": `variable storage_class {type = "string"}
                                  variable name {type = "string"}
                                  variable labels {type = "string"}`,
                    "outputs": `output bucket_name {value = "${var.name}"}`,
                },
                Outputs: []broker.BrokerVariable{{FieldName: "bucket_name"}},
            },
            ErrContains: "",
        },
        "extra inputs okay": {
            Action: TfServiceDefinitionV1Action{
                PlanInputs: []broker.BrokerVariable{{FieldName: "storage_class"}},
                UserInputs: []broker.BrokerVariable{{FieldName: "name"}},
                Computed:   []varcontext.DefaultVariable{{Name: "labels"}},
                Templates: map[string]string {
                    "variables": `variable storage_class {type = "string"}`,
                },
            },
            ErrContains: "",
        },
        "missing inputs": {
            Action: TfServiceDefinitionV1Action{
                PlanInputs: []broker.BrokerVariable{{FieldName: "storage_class"}},
                UserInputs: []broker.BrokerVariable{{FieldName: "name"}},
                Computed:   []varcontext.DefaultVariable{{Name: "labels"}},
                Templates: map[string]string {
                    "variables": `variable storage_class {type = "string"}
                                  variable not_defined {type = "string"}`,
                },
            },
            ErrContains: "fields used but not declared: template.not_defined",
        },

        "extra template outputs": {
            Action: TfServiceDefinitionV1Action{
                Templates: map[string]string {
                    "outputs": `output storage_class {value = "${var.name}"}
                                output name {value = "${var.name}"}
                                output labels {value = "${var.name}"}
                                output bucket_name {value = "${var.name}"}`,
                },
                Outputs: []broker.BrokerVariable{{FieldName: "bucket_name"}},
            },
            ErrContains: "template outputs [bucket_name labels name storage_class] must match declared outputs [bucket_name]:",
        },

        "missing template outputs": {
            Action: TfServiceDefinitionV1Action{
                Templates: map[string]string {
                    "outputs": ``,
                },
                Outputs: []broker.BrokerVariable{{FieldName: "bucket_name"}},
            },
            ErrContains: "template outputs [] must match declared outputs [bucket_name]:",
        },
    }

    for tn, tc := range cases {
        t.Run(tn, func(t *testing.T) {
            err := tc.Action.ValidateTemplateIO()
            if err == nil {
                if tc.ErrContains == "" {
                    return
                }

                t.Fatalf("Expected error to contain %q, got: <nil>", tc.ErrContains)
            } else {
                if tc.ErrContains == "" {
                    t.Fatalf("Expected no error, got: %v", err)
                }

                if !strings.Contains(err.Error(), tc.ErrContains) {
                    t.Fatalf("Expected error to contain %q, got: %v", tc.ErrContains, err)
                }
            }
        })
    }
}

func TestNewExampleTfServiceDefinition(t *testing.T) {
    example := NewExampleTfServiceDefinition()

    if err := example.Validate(); err != nil {
        t.Fatalf("example service definition should be valid, but got error: %v", err)
    }
}

func TestTfServiceDefinitionV1Plan_ToPlan(t *testing.T) {
    cases := map[string]struct {
        Definition TfServiceDefinitionV1Plan
        Expected   broker.ServicePlan
    }{
        "full": {
            Definition: TfServiceDefinitionV1Plan{
                Id:          "00000000-0000-0000-0000-000000000001",
                Name:        "example-email-plan",
                DisplayName: "example.com email builder",
                Description: "Builds emails for example.com.",
                Bullets:     []string{"information point 1", "information point 2", "some caveat here"},
                Free:        false,
                Properties: map[string]interface{}{
                    "domain": "example.com",
                },
            },
            Expected: broker.ServicePlan{
                ServicePlan: brokerapi.ServicePlan{
                    ID:          "00000000-0000-0000-0000-000000000001",
                    Name:        "example-email-plan",
                    Description: "Builds emails for example.com.",
                    Free:        brokerapi.FreeValue(false),
                    Metadata: &brokerapi.ServicePlanMetadata{
                        Bullets:     []string{"information point 1", "information point 2", "some caveat here"},
                        DisplayName: "example.com email builder",
                    },
                },
                ServiceProperties: map[string]interface{}{"domain": "example.com"}},
        },
    }

    for tn, tc := range cases {
        t.Run(tn, func(t *testing.T) {
            actual := tc.Definition.ToPlan()
            if !reflect.DeepEqual(actual, tc.Expected) {
                t.Fatalf("Expected: %v Actual: %v", tc.Expected, actual)
            }
        })
    }
}

func TestTfServiceDefinitionV1_ToService(t *testing.T) {
    definition := TfServiceDefinitionV1{
        Version:     1,
        Id:          "d34705c8-3edf-4ab8-93b3-d97f080da24c",
        Name:        "my-service-name",
        Description: "my-service-description",
        DisplayName: "My Service Name",

        ImageUrl:         "https://example.com/image.png",
        SupportUrl:       "https://example.com/support",
        DocumentationUrl: "https://example.com/docs",
        Plans:            []TfServiceDefinitionV1Plan{},
        RequiredEnvVars: []string{"EXAMPLE_ENV_VAR"},

        ProvisionSettings: TfServiceDefinitionV1Action{
            PlanInputs: []broker.BrokerVariable{
                {
                    FieldName: "plan-input-provision",
                    Type:      "string",
                    Details:   "description",
                },
            },
            UserInputs: []broker.BrokerVariable{
                {
                    FieldName: "user-input-provision",
                    Type:      "string",
                    Details:   "description",
                },
            },
            Computed: []varcontext.DefaultVariable{{Name: "computed-input-provision", Default: ""}},
            TemplateRef: "testdata/provision.tf",
        },

        BindSettings: TfServiceDefinitionV1Action{
            PlanInputs: []broker.BrokerVariable{
                {
                    FieldName: "plan-input-bind",
                    Type:      "integer",
                    Details:   "description",
                },
            },
            UserInputs: []broker.BrokerVariable{
                {
                    FieldName: "user-input-bind",
                    Type:      "string",
                    Details:   "description",
                },
            },
            Computed: []varcontext.DefaultVariable{{Name: "computed-input-bind", Default: ""}},
        },

        Examples: []broker.ServiceExample{},
    }

    service, err := definition.ToService(nil)
    if err == nil {
        t.Fatal(fmt.Errorf("Expected 'missing required env var EXAMPLE_ENV_VAR"))
    }
    if err.Error() != "missing required env var EXAMPLE_ENV_VAR" {
        t.Fatal(err)
    }

    os.Setenv("EXAMPLE_ENV_VAR", "example value")
    defer os.Unsetenv("EXAMPLE_ENV_VAR")
    service, err = definition.ToService(nil)
    if err != nil {
        t.Fatal(err)
    }

    expectEqual := func(field string, expected, actual interface{}) {
        if !reflect.DeepEqual(expected, actual) {
            t.Errorf("Expected %q to be equal. Expected: %#v, Actual: %#v", field, expected, actual)
        }
    }

    t.Run("basic-info", func(t *testing.T) {
        expectEqual("Id", definition.Id, service.Id)
        expectEqual("Name", definition.Name, service.Name)
        expectEqual("Description", definition.Description, service.Description)
        expectEqual("Bindable", true, service.Bindable)
        expectEqual("PlanUpdateable", false, service.PlanUpdateable)
        expectEqual("DisplayName", definition.DisplayName, service.DisplayName)
        expectEqual("DocumentationUrl", definition.DocumentationUrl, service.DocumentationUrl)
        expectEqual("SupportUrl", definition.SupportUrl, service.SupportUrl)
        expectEqual("ImageUrl", definition.ImageUrl, service.ImageUrl)
        expectEqual("Tags", definition.Tags, service.Tags)
    })

    t.Run("vars", func(t *testing.T) {
        expectEqual("ProvisionInputVariables", definition.ProvisionSettings.UserInputs, service.ProvisionInputVariables)
        expectEqual("ProvisionComputedVariables", []varcontext.DefaultVariable{
            {
                Name:      "computed-input-provision",
                Default:   "",
                Overwrite: false,
            },
            {
                Name:      "tf_id",
                Default:   "tf:${request.instance_id}:",
                Overwrite: true,
            },
        }, service.ProvisionComputedVariables)
        expectEqual("PlanVariables", append(definition.ProvisionSettings.PlanInputs, definition.BindSettings.PlanInputs...), service.PlanVariables)
        expectEqual("BindInputVariables", definition.BindSettings.UserInputs, service.BindInputVariables)
        expectEqual("BindComputedVariables", []varcontext.DefaultVariable{
            {Name: "plan-input-bind", Default: "${request.plan_properties[\"plan-input-bind\"]}", Overwrite: true, Type: "integer"},
            {Name: "computed-input-bind", Default: "", Overwrite: false, Type: ""},
            {Name: "tf_id", Default: "tf:${request.instance_id}:${request.binding_id}", Overwrite: true, Type: ""},
        }, service.BindComputedVariables)
        expectEqual("BindOutputVariables", append(definition.ProvisionSettings.Outputs, definition.BindSettings.Outputs...), service.BindOutputVariables)
    })

    t.Run("examples", func(t *testing.T) {
        expectEqual("Examples", definition.Examples, service.Examples)
    })

    t.Run("provider-builder", func(t *testing.T) {
        if service.ProviderBuilder == nil {
            t.Fatal("Expected provider builder to not be nil")
        }
    })
}
