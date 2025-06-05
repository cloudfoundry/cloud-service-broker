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
	"errors"
	"fmt"
	"reflect"
	"testing"

	"code.cloudfoundry.org/brokerapi/v13/domain"
	"github.com/spf13/viper"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/paramparser"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/validation"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/varcontext"
)

func ExampleServiceDefinition_UserDefinedPlansProperty() {
	service := ServiceDefinition{
		ID:   "00000000-0000-0000-0000-000000000000",
		Name: "left-handed-smoke-sifter",
	}

	fmt.Println(service.UserDefinedPlansProperty())

	// Output: service.left-handed-smoke-sifter.plans
}

func ExampleServiceDefinition_IsRoleWhitelistEnabled() {
	service := ServiceDefinition{
		ID:                   "00000000-0000-0000-0000-000000000000",
		Name:                 "left-handed-smoke-sifter",
		DefaultRoleWhitelist: []string{"a", "b", "c"},
	}
	fmt.Println(service.IsRoleWhitelistEnabled())

	service.DefaultRoleWhitelist = nil
	fmt.Println(service.IsRoleWhitelistEnabled())

	// Output: true
	// false
}

func ExampleServiceDefinition_TileUserDefinedPlansVariable() {
	service := ServiceDefinition{
		ID:   "00000000-0000-0000-0000-000000000000",
		Name: "google-spanner",
	}

	fmt.Println(service.TileUserDefinedPlansVariable())

	// Output: SPANNER_CUSTOM_PLANS
}

func ExampleServiceDefinition_GetPlanByID() {
	service := ServiceDefinition{
		ID:   "00000000-0000-0000-0000-000000000000",
		Name: "left-handed-smoke-sifter",
		Plans: []ServicePlan{
			{ServicePlan: domain.ServicePlan{ID: "test-plan", Name: "Builtin!"}},
		},
	}

	plan, err := service.GetPlanByID("test-plan")
	fmt.Printf("test-plan: %q %v\n", plan.Name, err)

	_, err = service.GetPlanByID("missing-plan")
	fmt.Printf("missing-plan: %s\n", err)

	// Output: test-plan: "Builtin!" <nil>
	// missing-plan: plan ID "missing-plan" could not be found
}

func TestServiceDefinition_CatalogEntry(t *testing.T) {
	cases := map[string]struct {
		UserPlans   any
		PlanIds     map[string]bool
		ExpectError bool
	}{
		"no-customization": {
			UserPlans:   nil,
			PlanIds:     map[string]bool{},
			ExpectError: false,
		},
		"custom-plans": {
			UserPlans:   `[{"id":"aaa","name":"aaa"},{"id":"bbb","name":"bbb"}]`,
			PlanIds:     map[string]bool{"aaa": true, "bbb": true},
			ExpectError: false,
		},
		"bad-plan-json": {
			UserPlans:   `333`,
			PlanIds:     map[string]bool{},
			ExpectError: true,
		},
	}

	service := ServiceDefinition{
		ID:   "00000000-0000-0000-0000-000000000000",
		Name: "left-handed-smoke-sifter",
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			viper.Set(service.UserDefinedPlansProperty(), tc.UserPlans)
			defer viper.Reset()

			plans, err := service.UserDefinedPlans(nil)
			hasErr := err != nil
			if hasErr != tc.ExpectError {
				t.Errorf("Expected Error? %v, got error: %v", tc.ExpectError, err)
			}

			if err == nil && len(plans) != len(tc.PlanIds) {
				t.Errorf("Expected %d plans, but got %d (%+v)", len(tc.PlanIds), len(plans), plans)

				for _, plan := range plans {
					if _, ok := tc.PlanIds[plan.ID]; !ok {
						t.Errorf("Got unexpected plan id %s, expected %+v", plan.ID, tc.PlanIds)
					}
				}
			}
		})
	}
}

func ExampleServiceDefinition_CatalogEntry() {
	service := ServiceDefinition{
		ID:   "00000000-0000-0000-0000-000000000000",
		Name: "left-handed-smoke-sifter",
		Plans: []ServicePlan{
			{ServicePlan: domain.ServicePlan{ID: "builtin-plan", Name: "Builtin!"}},
		},
		ProvisionInputVariables: []BrokerVariable{
			{FieldName: "location", Type: JSONTypeString, Default: "us"},
		},
		BindInputVariables: []BrokerVariable{
			{FieldName: "name", Type: JSONTypeString, Default: "name"},
		},
	}

	srvc := service.CatalogEntry()

	// Schemas should be nil by default
	fmt.Println("schemas with flag off:", srvc.ToPlain().Plans[0].Schemas)

	viper.Set("compatibility.enable-catalog-schemas", true)
	defer viper.Reset()

	srvc = service.CatalogEntry()

	eq := reflect.DeepEqual(srvc.ToPlain().Plans[0].Schemas, service.createSchemas())

	fmt.Println("schema was generated?", eq)

	// Output: schemas with flag off: <nil>
	// schema was generated? true
}

func TestServiceDefinition_ProvisionVariables(t *testing.T) {
	service := ServiceDefinition{
		ID:   "00000000-0000-0000-0000-000000000000",
		Name: "left-handed-smoke-sifter",
		Plans: []ServicePlan{
			{ServicePlan: domain.ServicePlan{ID: "builtin-plan", Name: "Builtin!"}},
		},
		ProvisionInputVariables: []BrokerVariable{
			{
				FieldName: "location",
				Type:      JSONTypeString,
				Default:   "us", // 7
			},
			{
				FieldName: "name",
				Type:      JSONTypeString,
				Default:   "name-${location}", // 7
				Constraints: validation.NewConstraintBuilder().
					MaxLength(30).
					Build(),
			},
		},
		ProvisionComputedVariables: []varcontext.DefaultVariable{
			{
				Name:      "location",
				Default:   "${str.truncate(10, location)}", // 1
				Overwrite: true,
			},
			{
				Name:      "maybe-missing",
				Default:   "default",
				Overwrite: false,
			},
			{
				Name:      "osb_context",
				Default:   `${json.marshal(request.context)}`,
				Overwrite: true,
				Type:      "object",
			},
			{
				Name:      "originatingIdentity",
				Default:   `${json.marshal(request.x_broker_api_originating_identity)}`,
				Overwrite: true,
				Type:      "object",
			},
		},
	}

	cases := map[string]struct {
		RawContext          string
		OriginatingIdentity map[string]any
		// precedence order - lowest number should win
		UserParams         string         // 4
		ProvisionOverrides map[string]any // 3
		ServiceProperties  map[string]any // 2
		DefaultOverride    string         // 5
		GlobalDefaults     string         // 6
		ExpectedError      error
		ExpectedContext    map[string]any

		// Some config values were historically expected to be provided as a string.
		// That is because these values were sourced from ENV Vars. But these values can
		// be provided via a config file. This flag is used to test the behavior of the
		// service definition when the config values are provided as a JSON object instead of a string.
		IsJSONFormat bool
	}{
		"empty": {
			UserParams:        "",
			ServiceProperties: map[string]any{},
			ExpectedContext: map[string]any{
				"location":            "us",
				"name":                "name-us",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"includes request context information": {
			UserParams: "",
			RawContext: `{"platform": "cloudfoundry", "organization_name": "acceptance"}`,
			OriginatingIdentity: map[string]any{
				"platform": "cloudfoundry",
				"value": map[string]string{
					"user_id": "683ea748-3092-4ff4-b656-39cacc4d5360",
				}},
			ServiceProperties: map[string]any{},
			ExpectedContext: map[string]any{
				"location":      "us",
				"name":          "name-us",
				"maybe-missing": "default",
				"osb_context": map[string]any{
					"platform":          "cloudfoundry",
					"organization_name": "acceptance",
				},
				"originatingIdentity": map[string]any{
					"platform": "cloudfoundry",
					"value":    map[string]any{"user_id": "683ea748-3092-4ff4-b656-39cacc4d5360"},
				},
			},
		},
		"service has missing param": {
			ServiceProperties: map[string]any{"maybe-missing": "custom"}, // 2
			UserParams:        "",                                        // 4
			ExpectedContext: map[string]any{
				"location":            "us",
				"name":                "name-us",
				"maybe-missing":       "custom",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"location gets truncated": {
			ServiceProperties: map[string]any{},                    // 2
			UserParams:        `{"location": "averylonglocation"}`, // 4
			ExpectedContext: map[string]any{
				"location":            "averylongl",
				"name":                "name-averylonglocation",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"user location and name": {
			ServiceProperties: map[string]any{},                   // 2
			UserParams:        `{"location": "eu", "name":"foo"}`, // 4
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "foo",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"user tries to overwrite service var": {
			ServiceProperties: map[string]any{"service-provided": "custom"},                  // 2
			UserParams:        `{"location": "eu", "name":"foo", "service-provided":"test"}`, // 4
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "foo",
				"maybe-missing":       "default",
				"service-provided":    "custom",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"operator defaults override computed defaults": {
			ServiceProperties: map[string]any{},    // 2
			UserParams:        "",                  // 4
			DefaultOverride:   `{"location":"eu"}`, // 5
			GlobalDefaults:    `{"location":"az"}`, // 6
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "name-eu",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"user values override operator defaults": {
			ServiceProperties: map[string]any{},    // 2
			UserParams:        `{"location":"nz"}`, // 4
			DefaultOverride:   `{"location":"eu"}`, // 5
			GlobalDefaults:    "{}",
			ExpectedContext: map[string]any{
				"location":            "nz",
				"name":                "name-nz",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"operator defaults are not evaluated": {
			ServiceProperties: map[string]any{},             // 2
			UserParams:        `{"location":"us"}`,          // 4
			DefaultOverride:   `{"name":"foo-${location}"}`, // 5
			GlobalDefaults:    "{}",
			ExpectedContext: map[string]any{
				"location":            "us",
				"name":                "foo-${location}",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"invalid-request": {
			UserParams:    `{"name":"some-name-that-is-longer-than-thirty-characters"}`,
			ExpectedError: errors.New("1 error(s) occurred: name: String length must be less than or equal to 30"),
		},
		"provision_overrides override user params and global_defaults but not computed defaults": {
			ServiceProperties:  map[string]any{},                 // 2
			ProvisionOverrides: map[string]any{"location": "eu"}, // 3
			UserParams:         `{"location":"us"}`,              // 4
			DefaultOverride:    "{}",                             // 5
			GlobalDefaults:     `{"location":"az"}`,              // 6
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "name-eu",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"global_default override defaults but not computed defaults": {
			ServiceProperties:  map[string]any{},    // 2
			ProvisionOverrides: map[string]any{},    // 3
			UserParams:         "{}",                // 4
			DefaultOverride:    "{}",                // 5
			GlobalDefaults:     `{"location":"az"}`, // 6
			ExpectedContext: map[string]any{
				"location":            "az",
				"name":                "name-az",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"bogus global default json": {
			ServiceProperties:  map[string]any{},    // 2
			ProvisionOverrides: map[string]any{},    // 3
			UserParams:         "{}",                // 4
			DefaultOverride:    "{}",                // 5
			GlobalDefaults:     `{"location","az"}`, // 6
			ExpectedContext: map[string]any{
				"location":            "az",
				"name":                "name-az",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
			ExpectedError: fmt.Errorf("failed unmarshaling config value provision.defaults"),
		},
		"provision_overrides override user params and global_defaults but not computed defaults using JSON objects": {
			ServiceProperties:  map[string]any{},                 // 2
			ProvisionOverrides: map[string]any{"location": "eu"}, // 3
			UserParams:         `{"location":"us"}`,              // 4
			DefaultOverride:    "{}",                             // 5
			GlobalDefaults:     `{"location":"az"}`,              // 6
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "name-eu",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
			IsJSONFormat: true,
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			setViperConfig(service.ProvisionDefaultOverrideProperty(), tc.DefaultOverride, tc.IsJSONFormat)
			setViperConfig(GlobalProvisionDefaults, tc.GlobalDefaults, tc.IsJSONFormat)

			defer viper.Reset()

			details := paramparser.ProvisionDetails{
				RequestParams:  mustUnmarshal(tc.UserParams),
				RequestContext: mustUnmarshal(tc.RawContext),
			}

			plan := ServicePlan{ServiceProperties: tc.ServiceProperties, ProvisionOverrides: tc.ProvisionOverrides}
			vars, err := service.ProvisionVariables("instance-id-here", details, plan, tc.OriginatingIdentity)

			expectError(t, tc.ExpectedError, err)

			if tc.ExpectedError == nil && !reflect.DeepEqual(vars.ToMap(), tc.ExpectedContext) {
				t.Errorf("Expected context: %v got %v", tc.ExpectedContext, vars.ToMap())
			}
		})
	}
}

// Function to set viper configuration based on JSON format flag.
func setViperConfig(propertyName string, value string, isJSONFormat bool) {
	if len(value) == 0 {
		return
	}

	if isJSONFormat {
		viper.Set(propertyName, mustUnmarshal(value))
	} else {
		viper.Set(propertyName, value)
	}
}

func TestServiceDefinition_UpdateVariables(t *testing.T) {
	service := ServiceDefinition{
		ID:   "00000000-0000-0000-0000-000000000000",
		Name: "left-handed-smoke-sifter",
		Plans: []ServicePlan{
			{ServicePlan: domain.ServicePlan{ID: "builtin-plan", Name: "Builtin!"}},
		},
		ProvisionInputVariables: []BrokerVariable{
			{
				FieldName: "location",
				Type:      JSONTypeString,
				Default:   "us", // 7
			},
			{
				FieldName: "name",
				Type:      JSONTypeString,
				Default:   "name-${location}", // 7
				Constraints: validation.NewConstraintBuilder().
					MaxLength(30).
					Build(),
			},
		},
		ProvisionComputedVariables: []varcontext.DefaultVariable{
			{
				Name:      "location",
				Default:   "${str.truncate(10, location)}", // 1
				Overwrite: true,
			},
			{
				Name:      "maybe-missing",
				Default:   "default",
				Overwrite: false,
			},
			{
				Name:      "osb_context",
				Default:   `${json.marshal(request.context)}`,
				Overwrite: true,
				Type:      "object",
			},
			{
				Name:      "originatingIdentity",
				Default:   `${json.marshal(request.x_broker_api_originating_identity)}`,
				Overwrite: true,
				Type:      "object",
			},
		},
		GlobalLabels: map[string]string{"key1": "value1", "key2": "value2"},
	}

	cases := map[string]struct {
		RawContext          string
		OriginatingIdentity map[string]any
		// precedence order - lowest number should win
		MergedUserProvidedParams string         // 4
		ProvisionOverrides       map[string]any // 3
		ServiceProperties        map[string]any // 2
		DefaultOverride          string         // 5
		GlobalDefaults           string         // 6
		ExpectedError            error
		ExpectedContext          map[string]any
	}{
		"empty": {
			ServiceProperties: map[string]any{},
			ExpectedContext: map[string]any{
				"location":            "us",
				"name":                "name-us",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"includes request context information": {
			RawContext: `{"platform": "cloudfoundry", "organization_name": "acceptance"}`,
			OriginatingIdentity: map[string]any{
				"platform": "cloudfoundry",
				"value": map[string]string{
					"user_id": "683ea748-3092-4ff4-b656-39cacc4d5360",
				}},
			ServiceProperties: map[string]any{},
			ExpectedContext: map[string]any{
				"location":      "us",
				"name":          "name-us",
				"maybe-missing": "default",
				"osb_context": map[string]any{
					"platform":          "cloudfoundry",
					"organization_name": "acceptance",
				},
				"originatingIdentity": map[string]any{
					"platform": "cloudfoundry",
					"value":    map[string]any{"user_id": "683ea748-3092-4ff4-b656-39cacc4d5360"},
				},
			},
		},
		"service has missing param": {
			ServiceProperties: map[string]any{"maybe-missing": "custom"}, // 2
			ExpectedContext: map[string]any{
				"location":            "us",
				"name":                "name-us",
				"maybe-missing":       "custom",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"location gets truncated": {
			ServiceProperties:        map[string]any{},                    // 2
			MergedUserProvidedParams: `{"location": "averylonglocation"}`, // 4
			ExpectedContext: map[string]any{
				"location":            "averylongl",
				"name":                "name-averylonglocation",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"user location and name": {
			ServiceProperties:        map[string]any{},                   // 2
			MergedUserProvidedParams: `{"location": "eu", "name":"foo"}`, // 4
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "foo",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"user tries to overwrite service var": {
			ServiceProperties:        map[string]any{"service-provided": "custom"},                  // 2
			MergedUserProvidedParams: `{"location": "eu", "name":"foo", "service-provided":"test"}`, // 4
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "foo",
				"maybe-missing":       "default",
				"service-provided":    "custom",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"operator defaults override computed defaults": {
			ServiceProperties: map[string]any{},    // 2
			DefaultOverride:   `{"location":"eu"}`, // 5
			GlobalDefaults:    `{"location":"az"}`, // 6
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "name-eu",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"user values override operator defaults": {
			ServiceProperties:        map[string]any{},    // 2
			MergedUserProvidedParams: `{"location":"nz"}`, // 4
			DefaultOverride:          `{"location":"eu"}`, // 5
			GlobalDefaults:           "{}",
			ExpectedContext: map[string]any{
				"location":            "nz",
				"name":                "name-nz",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"operator defaults are not evaluated": {
			ServiceProperties:        map[string]any{},             // 2
			MergedUserProvidedParams: `{"location":"us"}`,          // 4
			DefaultOverride:          `{"name":"foo-${location}"}`, // 5
			GlobalDefaults:           "{}",
			ExpectedContext: map[string]any{
				"location":            "us",
				"name":                "foo-${location}",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"invalid-request": {
			MergedUserProvidedParams: `{"name":"some-name-that-is-longer-than-thirty-characters"}`,
			ExpectedError:            errors.New("1 error(s) occurred: name: String length must be less than or equal to 30"),
		},
		"provision_overrides override user params and global_defaults but not computed defaults": {
			ServiceProperties:        map[string]any{},                 // 2
			ProvisionOverrides:       map[string]any{"location": "eu"}, // 3
			MergedUserProvidedParams: `{"location":"us"}`,              // 4
			DefaultOverride:          "{}",                             // 5
			GlobalDefaults:           `{"location":"az"}`,              // 6
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "name-eu",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"global_default override defaults but not computed defaults": {
			ServiceProperties:        map[string]any{},    // 2
			ProvisionOverrides:       map[string]any{},    // 3
			MergedUserProvidedParams: "{}",                // 4
			DefaultOverride:          "{}",                // 5
			GlobalDefaults:           `{"location":"az"}`, // 6
			ExpectedContext: map[string]any{
				"location":            "az",
				"name":                "name-az",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"bogus global default json": {
			ServiceProperties:        map[string]any{},    // 2
			ProvisionOverrides:       map[string]any{},    // 3
			MergedUserProvidedParams: "{}",                // 4
			DefaultOverride:          "{}",                // 5
			GlobalDefaults:           `{"location","az"}`, // 6
			ExpectedContext: map[string]any{
				"location":            "az",
				"name":                "name-az",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
			ExpectedError: fmt.Errorf("failed unmarshaling config value provision.defaults"),
		},
		"provision location and name": {
			ServiceProperties:        map[string]any{},                   // 2
			MergedUserProvidedParams: `{"location": "eu", "name":"foo"}`, // 5
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "foo",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"update location and name": {
			ServiceProperties:        map[string]any{},                      // 2
			MergedUserProvidedParams: `{"name":"update", "location": "eu"}`, // 5
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "update",
				"maybe-missing":       "default",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			if len(tc.DefaultOverride) > 0 {
				viper.Set(service.ProvisionDefaultOverrideProperty(), tc.DefaultOverride)
			}
			if len(tc.GlobalDefaults) > 0 {
				viper.Set(GlobalProvisionDefaults, tc.GlobalDefaults)
			}
			defer viper.Reset()

			details := paramparser.UpdateDetails{RequestContext: mustUnmarshal(tc.RawContext)}
			mergedUserProvidedParams := mustUnmarshal(tc.MergedUserProvidedParams)
			plan := ServicePlan{ServiceProperties: tc.ServiceProperties, ProvisionOverrides: tc.ProvisionOverrides}
			vars, err := service.UpdateVariables("instance-id-here", details, mergedUserProvidedParams, plan, tc.OriginatingIdentity)

			expectError(t, tc.ExpectedError, err)

			if tc.ExpectedError == nil && !reflect.DeepEqual(vars.ToMap(), tc.ExpectedContext) {
				t.Errorf("Expected context: %v got %v", tc.ExpectedContext, vars.ToMap())
			}
		})
	}
}

func TestServiceDefinition_BindVariables(t *testing.T) {
	service := ServiceDefinition{
		ID:   "00000000-0000-0000-0000-000000000000",
		Name: "left-handed-smoke-sifter",
		Plans: []ServicePlan{
			{
				ServicePlan: domain.ServicePlan{
					ID:   "builtin-plan",
					Name: "Builtin!",
				},
				ServiceProperties: map[string]any{
					"service-property": "operator-set",
				},
			},
		},
		BindInputVariables: []BrokerVariable{
			{
				FieldName: "location",
				Type:      JSONTypeString,
				Default:   "us",
			},
			{
				FieldName: "name",
				Type:      JSONTypeString,
				Default:   "name-${location}",
				Constraints: validation.NewConstraintBuilder().
					MaxLength(30).
					Build(),
			},
		},
		BindComputedVariables: []varcontext.DefaultVariable{
			{
				Name:      "location",
				Default:   "${str.truncate(10, location)}",
				Overwrite: true,
			},
			{
				Name:      "instance-foo",
				Default:   `${instance.details["foo"]}`,
				Overwrite: true,
			},
			{
				Name:      "service-prop",
				Default:   `${request.plan_properties["service-property"]}`,
				Overwrite: true,
			},
			{
				Name:      "osb_context",
				Default:   `${json.marshal(request.context)}`,
				Overwrite: true,
				Type:      "object",
			},
			{
				Name:      "originatingIdentity",
				Default:   `${json.marshal(request.x_broker_api_originating_identity)}`,
				Overwrite: true,
				Type:      "object",
			},
		},
	}

	cases := map[string]struct {
		UserParams          string
		DefaultOverride     string
		BindOverrides       map[string]any
		ExpectedError       error
		ExpectedContext     map[string]any
		InstanceVars        map[string]any
		RawContext          string
		OriginatingIdentity map[string]any
	}{
		"empty": {
			UserParams:   "",
			InstanceVars: map[string]any{"foo": ""},
			ExpectedContext: map[string]any{
				"location":            "us",
				"name":                "name-us",
				"instance-foo":        "",
				"service-prop":        "operator-set",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"includes request context information": {
			UserParams: "",
			RawContext: `{"platform": "cloudfoundry", "organization_name": "acceptance"}`,
			OriginatingIdentity: map[string]any{
				"platform": "cloudfoundry",
				"value": map[string]string{
					"user_id": "683ea748-3092-4ff4-b656-39cacc4d5360",
				}},
			InstanceVars: map[string]any{"foo": ""},
			ExpectedContext: map[string]any{
				"location":     "us",
				"name":         "name-us",
				"instance-foo": "",
				"service-prop": "operator-set",
				"osb_context": map[string]any{
					"platform":          "cloudfoundry",
					"organization_name": "acceptance",
				},
				"originatingIdentity": map[string]any{
					"platform": "cloudfoundry",
					"value":    map[string]any{"user_id": "683ea748-3092-4ff4-b656-39cacc4d5360"},
				},
			},
		},
		"location gets truncated": {
			UserParams:   `{"location": "averylonglocation"}`,
			InstanceVars: map[string]any{"foo": "default"},
			ExpectedContext: map[string]any{
				"location":            "averylongl",
				"name":                "name-averylonglocation",
				"instance-foo":        "default",
				"service-prop":        "operator-set",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"user location and name": {
			UserParams:   `{"location": "eu", "name":"foo"}`,
			InstanceVars: map[string]any{"foo": "default"},
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "foo",
				"instance-foo":        "default",
				"service-prop":        "operator-set",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"operator defaults override computed defaults": {
			UserParams:      "",
			InstanceVars:    map[string]any{"foo": "default"},
			DefaultOverride: `{"location":"eu"}`,
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "name-eu",
				"instance-foo":        "default",
				"service-prop":        "operator-set",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"user values override operator defaults": {
			UserParams:      `{"location":"nz"}`,
			InstanceVars:    map[string]any{"foo": "default"},
			DefaultOverride: `{"location":"eu"}`,
			ExpectedContext: map[string]any{
				"location":            "nz",
				"name":                "name-nz",
				"instance-foo":        "default",
				"service-prop":        "operator-set",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"operator defaults are not evaluated": {
			UserParams:      `{"location":"us"}`,
			InstanceVars:    map[string]any{"foo": "default"},
			DefaultOverride: `{"name":"foo-${location}"}`,
			ExpectedContext: map[string]any{
				"location":            "us",
				"name":                "foo-${location}",
				"instance-foo":        "default",
				"service-prop":        "operator-set",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"instance info can get parsed": {
			UserParams:   `{"location":"us"}`,
			InstanceVars: map[string]any{"foo": "bar"},
			ExpectedContext: map[string]any{
				"location":            "us",
				"name":                "name-us",
				"instance-foo":        "bar",
				"service-prop":        "operator-set",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
		"invalid-request": {
			UserParams:    `{"name":"some-name-that-is-longer-than-thirty-characters"}`,
			InstanceVars:  map[string]any{"foo": ""},
			ExpectedError: errors.New("1 error(s) occurred: name: String length must be less than or equal to 30"),
		},
		"bind_overrides override user params but not computed defaults": {
			UserParams:    `{"location":"us"}`,
			InstanceVars:  map[string]any{"foo": "default"},
			BindOverrides: map[string]any{"location": "eu"},
			ExpectedContext: map[string]any{
				"location":            "eu",
				"name":                "name-eu",
				"instance-foo":        "default",
				"service-prop":        "operator-set",
				"osb_context":         map[string]any{},
				"originatingIdentity": map[string]any{},
			},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			viper.Set(service.BindDefaultOverrideProperty(), tc.DefaultOverride)
			defer viper.Reset()

			details := domain.BindDetails{RawParameters: json.RawMessage(tc.UserParams), RawContext: json.RawMessage(tc.RawContext), AppGUID: "fake-app-guid"}
			parsedDetails, err := paramparser.ParseBindDetails(details)
			if err != nil {
				t.Fatalf("failed to parse binding details: %s", err)
			}
			instance := storage.ServiceInstanceDetails{Outputs: tc.InstanceVars}

			service.Plans[0].BindOverrides = tc.BindOverrides
			vars, err := service.BindVariables(instance, "binding-id-here", parsedDetails, &service.Plans[0], tc.OriginatingIdentity)

			expectError(t, tc.ExpectedError, err)

			if tc.ExpectedError == nil && !reflect.DeepEqual(vars.ToMap(), tc.ExpectedContext) {
				t.Errorf("Expected context: %v got %v", tc.ExpectedContext, vars.ToMap())
			}
		})
	}
}

func TestServiceDefinition_createSchemas(t *testing.T) {
	service := ServiceDefinition{
		ID:   "00000000-0000-0000-0000-000000000000",
		Name: "left-handed-smoke-sifter",
		Plans: []ServicePlan{
			{ServicePlan: domain.ServicePlan{ID: "builtin-plan", Name: "Builtin!"}},
		},
		ProvisionInputVariables: []BrokerVariable{
			{FieldName: "location", Type: JSONTypeString, Default: "us"},
		},
		BindInputVariables: []BrokerVariable{
			{FieldName: "name", Type: JSONTypeString, Default: "name"},
		},
	}

	schemas := service.createSchemas()
	if schemas == nil {
		t.Fatal("Schemas was nil, expected non-nil value")
	}

	// it populates the instance create schema with the fields in ProvisionInputVariables
	instanceCreate := schemas.Instance.Create
	if instanceCreate.Parameters == nil {
		t.Error("instance create params were nil, expected a schema")
	}

	expectedCreateParams := CreateJSONSchema(service.ProvisionInputVariables)
	if !reflect.DeepEqual(instanceCreate.Parameters, expectedCreateParams) {
		t.Errorf("expected create params to be: %v got %v", expectedCreateParams, instanceCreate.Parameters)
	}

	// It leaves the instance update schema blank.
	instanceUpdate := schemas.Instance.Update
	if instanceUpdate.Parameters != nil {
		t.Error("instance update params were not nil, expected nil")
	}

	// it populates the binding create schema with the fields in BindInputVariables.
	bindCreate := schemas.Binding.Create
	if bindCreate.Parameters == nil {
		t.Error("bind create params were not nil, expected a schema")
	}

	expectedBindCreateParams := CreateJSONSchema(service.BindInputVariables)
	if !reflect.DeepEqual(bindCreate.Parameters, expectedBindCreateParams) {
		t.Errorf("expected create params to be: %v got %v", expectedBindCreateParams, bindCreate.Parameters)
	}
}

func expectError(t *testing.T, expected, actual error) {
	t.Helper()
	expectedErr := expected != nil
	gotErr := actual != nil

	switch {
	case expectedErr && gotErr:
		if expected.Error() != actual.Error() {
			t.Fatalf("Expected: %v, got: %v", expected, actual)
		}
	case expectedErr && !gotErr:
		t.Fatalf("Expected: %v, got: %v", expected, actual)
	case !expectedErr && gotErr:
		t.Fatalf("Expected no error but got: %v", actual)
	}
}

func mustUnmarshal(input string) (result map[string]any) {
	if len(input) == 0 {
		return
	}

	if err := json.Unmarshal([]byte(input), &result); err != nil {
		panic(err)
	}

	return
}
