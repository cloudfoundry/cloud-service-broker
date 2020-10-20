// Copyright 2020 the VMware, Inc.
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

package wrapper

import (
    "reflect"
    "strings"
    "testing"
    "unicode"
)

func compareIgnoreWhiteSpace(str1, str2 string) bool {
    return stripWhiteSpace(str1) == stripWhiteSpace(str2)
}

func stripWhiteSpace(str string) string {
    return strings.Map(func(r rune) rune {
        if unicode.IsSpace(r) {
            // if the character is a space, drop it
            return -1
        }
        // else keep it in the string
        return r
    }, str)
}

func TestTfImportTransform_CleanTf(t *testing.T) {
    cases := map[string]struct {
        transformer TfTransformer
        input       string
        expected    string
    }{
        "remove-id": {
            transformer: TfTransformer{
                ParametersToRemove: []string{"azurerm_mssql_database.azure_sql_db.id"},
            },
            input: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    creation_date                    = "2020-08-26T18:15:12.057Z"
    default_secondary_location       = "West US"
    edition                          = "Basic"
    id                               = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server/databases/db"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
    }
    zone_redundant                   = false

    threat_detection_policy {
        disabled_alerts      = []
        email_account_admins = "Disabled"
        email_addresses      = []
        retention_days       = 0
        state                = "Disabled"
        use_server_default   = "Disabled"
        id                   = "should be kept"
    }

    timeouts {}
}`,
            expected: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    creation_date                    = "2020-08-26T18:15:12.057Z"
    default_secondary_location       = "West US"
    edition                          = "Basic"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
    }
    zone_redundant                   = false

    threat_detection_policy {
        disabled_alerts      = []
        email_account_admins = "Disabled"
        email_addresses      = []
        retention_days       = 0
        state                = "Disabled"
        use_server_default   = "Disabled"
        id                   = "should be kept"		
    }

    timeouts {}
}`,
        },
        "remove-multiple": {
            transformer: TfTransformer{
                ParametersToRemove: []string{"azurerm_mssql_database.azure_sql_db.id",
                    "azurerm_mssql_database.azure_sql_db.creation_date",
                    "azurerm_mssql_database.azure_sql_db.default_secondary_location",
                    "azurerm_mssql_database.azure_sql_db.partner_servers.location",
                    "azurerm_mssql_database.azure_sql_db.partner_servers.role"},
            },
            input: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    edition                          = "Basic"
    id                               = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server/databases/db"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
    }
    zone_redundant                   = false

    partner_servers {
        id       = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-fog-subsume-test-server"
        location = "West US"
        role     = "Secondary"
    }
}`,
            expected: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    edition                          = "Basic"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
    }
    zone_redundant                   = false

    partner_servers {
        id       = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-fog-subsume-test-server"
    }
}`,
        },
        "remove-none": {
            transformer: TfTransformer{
                ParametersToRemove: []string{},
            },
            input: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    creation_date                    = "2020-08-26T18:15:12.057Z"
    default_secondary_location       = "West US"
    edition                          = "Basic"
    id                               = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server/databases/db"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
    }
    zone_redundant                   = false

    threat_detection_policy {
        disabled_alerts      = []
        email_account_admins = "Disabled"
        email_addresses      = []
        retention_days       = 0
        state                = "Disabled"
        use_server_default   = "Disabled"
        id                   = "should be kept"		
    }

    timeouts {}
}`,
            expected: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    creation_date                    = "2020-08-26T18:15:12.057Z"
    default_secondary_location       = "West US"
    edition                          = "Basic"
    id                               = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server/databases/db"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
    }
    zone_redundant                   = false

    threat_detection_policy {
        disabled_alerts      = []
        email_account_admins = "Disabled"
        email_addresses      = []
        retention_days       = 0
        state                = "Disabled"
        use_server_default   = "Disabled"
        id                   = "should be kept"		
    }

    timeouts {}
}`,
        },
    }

    for tn, tc := range cases {
        t.Run(tn, func(t *testing.T) {
            output := tc.transformer.CleanTf(tc.input)
            if !compareIgnoreWhiteSpace(output, tc.expected) {
                t.Fatalf("Expected %s, actual %s", tc.expected, output)
            }
        })
    }
}

func TestTfImportTransform_ReplaceParametersInTf(t *testing.T) {
    cases := map[string]struct {
        transformer        TfTransformer
        input              string
        expected           string
        expectedParameters map[string]string
    }{
        "none": {
            transformer: TfTransformer{
                ParameterMappings: []ParameterMapping{},
            },
            expectedParameters: map[string]string{},
            input: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    creation_date                    = "2020-08-26T18:15:12.057Z"
    default_secondary_location       = "West US"
    edition                          = "Basic"
    id                               = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server/databases/db"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
    }
    zone_redundant                   = false

    threat_detection_policy {
        disabled_alerts      = []
        email_account_admins = "Disabled"
        email_addresses      = []
        retention_days       = 0
        state                = "Disabled"
        use_server_default   = "Disabled"
    }

    timeouts {}
}`,
            expected: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    creation_date                    = "2020-08-26T18:15:12.057Z"
    default_secondary_location       = "West US"
    edition                          = "Basic"
    id                               = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server/databases/db"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
    }
    zone_redundant                   = false

    threat_detection_policy {
        disabled_alerts      = []
        email_account_admins = "Disabled"
        email_addresses      = []
        retention_days       = 0
        state                = "Disabled"
        use_server_default   = "Disabled"
    }

    timeouts {}
}`,
        },
        "edition": {
            transformer: TfTransformer{
                ParameterMappings: []ParameterMapping{
                    {
                        TfVariable:    "edition",
                        ParameterName: "var.edition",
                    },
                },
            },
            expectedParameters: map[string]string{
                "edition": "Basic",
            },
            input: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    creation_date                    = "2020-08-26T18:15:12.057Z"
    default_secondary_location       = "West US"
    edition                          = "Basic"
    id                               = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server/databases/db"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
    }
    zone_redundant                   = false

    threat_detection_policy {
        disabled_alerts      = []
        email_account_admins = "Disabled"
        email_addresses      = []
        retention_days       = 0
        state                = "Disabled"
        use_server_default   = "Disabled"
    }

    timeouts {}
}`,
            expected: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    creation_date                    = "2020-08-26T18:15:12.057Z"
    default_secondary_location       = "West US"
    edition                          = var.edition
    id                               = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server/databases/db"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
    }
    zone_redundant                   = false

    threat_detection_policy {
        disabled_alerts      = []
        email_account_admins = "Disabled"
        email_addresses      = []
        retention_days       = 0
        state                = "Disabled"
        use_server_default   = "Disabled"
    }

    timeouts {}
}`,
        },
        "tags": {
            transformer: TfTransformer{
                ParameterMappings: []ParameterMapping{
                    {
                        TfVariable:    "tags",
                        ParameterName: "var.labels",
                    },
                },
            },
            expectedParameters: map[string]string{
    //             "labels": `{
    //     "user-agent" = "meta-azure-service-broker"
    // }`,
            },
            input: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    creation_date                    = "2020-08-26T18:15:12.057Z"
    default_secondary_location       = "West US"
    edition                          = "Basic"
    id                               = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server/databases/db"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
    }
    zone_redundant                   = false

    threat_detection_policy {
        disabled_alerts      = []
        email_account_admins = "Disabled"
        email_addresses      = []
        retention_days       = 0
        state                = "Disabled"
        use_server_default   = "Disabled"
    }

    timeouts {}
}`,
            expected: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    creation_date                    = "2020-08-26T18:15:12.057Z"
    default_secondary_location       = "West US"
    edition                          = "Basic"
    id                               = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server/databases/db"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = var.labels
    zone_redundant                   = false

    threat_detection_policy {
        disabled_alerts      = []
        email_account_admins = "Disabled"
        email_addresses      = []
        retention_days       = 0
        state                = "Disabled"
        use_server_default   = "Disabled"
    }

    timeouts {}
}`,
		},
        "local": {
            transformer: TfTransformer{
                ParameterMappings: []ParameterMapping{
                    {
                        TfVariable:    "sku_name",
                        ParameterName: "local.sku_name",
                    },
                },
            },
            expectedParameters: map[string]string{
                "sku_name": "GP_Gen5_4",
            },
            input: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    creation_date                    = "2020-08-26T18:15:12.057Z"
    default_secondary_location       = "West US"
    edition                          = "Basic"
    id                               = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server/databases/db"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
    }
    zone_redundant                   = false

    threat_detection_policy {
        disabled_alerts      = []
        email_account_admins = "Disabled"
        email_addresses      = []
        retention_days       = 0
        state                = "Disabled"
        use_server_default   = "Disabled"
    }

    timeouts {}
	sku_name = "GP_Gen5_4"
}`,
            expected: `resource "azurerm_mssql_database" "azure_sql_db" {
    collation                        = "SQL_Latin1_General_CP1_CI_AS"
    creation_date                    = "2020-08-26T18:15:12.057Z"
    default_secondary_location       = "West US"
    edition                          = "Basic"
    id                               = "/subscriptions/899bf076-632b-4143-b015-43da8179e53f/resourceGroups/broker-cf-test/providers/Microsoft.Sql/servers/masb-subsume-test-server/databases/db"
    location                         = "eastus"
    max_size_bytes                   = "2147483648"
    name                             = "db"
    read_scale                       = false
    requested_service_objective_id   = "dd6d99bb-f193-4ec1-86f2-43d3bccbc49c"
    requested_service_objective_name = "Basic"
    resource_group_name              = "broker-cf-test"
    server_name                      = "masb-subsume-test-server"
    tags                             = {
        "user-agent" = "meta-azure-service-broker"
	}
	zone_redundant                   = false

    threat_detection_policy {
        disabled_alerts      = []
        email_account_admins = "Disabled"
        email_addresses      = []
        retention_days       = 0
        state                = "Disabled"
        use_server_default   = "Disabled"
    }

	timeouts {}
	sku_name = local.sku_name
}`,
        },				
    }

    for tn, tc := range cases {
        t.Run(tn, func(t *testing.T) {
            output, parameters, err := tc.transformer.ReplaceParametersInTf(tc.input)
            if err != nil {
                t.Fatal(err)
            }
            if !compareIgnoreWhiteSpace(output, tc.expected) {
                t.Fatalf("Expected %s, actual %s", tc.expected, output)
            }
            if !reflect.DeepEqual(parameters, tc.expectedParameters) {
                t.Fatalf("Expected %v, actual %v", tc.expectedParameters, parameters)
            }
        })
    }
}
