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

package cmd

import (
	"encoding/json"
	"log"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/client"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/featureflags"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/server"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils"
	"github.com/google/uuid"
	"github.com/pivotal-cf/brokerapi/v12/domain"
	"github.com/spf13/cobra"
)

var (
	serviceID      string
	planID         string
	instanceID     string
	bindingID      string
	parametersJSON string
	oldVersion     string
	newVersion     string

	serviceName     string
	exampleName     string
	fileName        string
	exampleJobCount int
)

func init() {

	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "A CLI client for the service broker",
		Long: `A CLI client for the service broker.

The client commands use the same configuration values as the server and operate
on localhost using the HTTP protocol.

Configuration Params:

 - api.user
 - api.password
 - api.port
 - api.hostname (default: localhost)

Environment Variables:

 - GSB_API_USER
 - GSB_API_PASSWORD
 - GSB_API_PORT
 - GSB_API_HOSTNAME

The client commands return formatted JSON when run if the exit code is 0:

	{
	    "url": "http://user:pass@localhost:8000/v2/catalog",
	    "http_method": "GET",
	    "status_code": 200,
	    "response": // Response Body as JSON
	}

Exit codes DO NOT correspond with status_code, if a request was made and the
response could be parsed then the exit code will be 0.
Non-zero exit codes indicate a failure in the executable.

Because of the format, you can use the client to do automated testing of your
user-defined plans.
`,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	rootCmd.AddCommand(clientCmd)

	clientCatalogCmd := newClientCommand("catalog", "Show the service catalog", func(client *client.Client) *client.BrokerResponse {
		return client.Catalog(uuid.NewString())
	})

	provisionCmd := newClientCommand("provision", "Provision a service", func(client *client.Client) *client.BrokerResponse {
		return client.Provision(instanceID, serviceID, planID, uuid.NewString(), json.RawMessage(parametersJSON))
	})

	deprovisionCmd := newClientCommand("deprovision", "Deprovision a service", func(client *client.Client) *client.BrokerResponse {
		return client.Deprovision(instanceID, serviceID, planID, uuid.NewString())
	})

	bindCmd := newClientCommand("bind", "Bind to a service", func(client *client.Client) *client.BrokerResponse {
		return client.Bind(instanceID, bindingID, serviceID, planID, uuid.NewString(), json.RawMessage(parametersJSON))
	})

	unbindCmd := newClientCommand("unbind", "Unbind a service", func(client *client.Client) *client.BrokerResponse {
		return client.Unbind(instanceID, bindingID, serviceID, planID, uuid.NewString())
	})

	lastCmd := newClientCommand("last", "Get the status of the last operation", func(client *client.Client) *client.BrokerResponse {
		return client.LastOperation(instanceID, uuid.NewString())
	})

	updateCmd := newClientCommand("update", "Update the instance details", func(client *client.Client) *client.BrokerResponse {
		return client.Update(instanceID, serviceID, planID, uuid.NewString(), json.RawMessage(parametersJSON), domain.PreviousValues{}, nil)
	})

	upgradeCmd := newClientCommand("upgrade", "Upgrade the service instance", func(c *client.Client) *client.BrokerResponse {
		return c.Update(instanceID, serviceID, planID, uuid.NewString(), json.RawMessage("{}"),
			domain.PreviousValues{ServiceID: serviceID, PlanID: planID, MaintenanceInfo: &domain.MaintenanceInfo{Version: oldVersion}},
			&domain.MaintenanceInfo{Version: newVersion})
	})

	examplesCmd := &cobra.Command{
		Use:   "examples",
		Short: "Display available examples",
		Long:  "Display available examples",
		Run: func(cmd *cobra.Command, args []string) {
			log.Printf("%s: %s\n\n", "example name", "service name")
			for _, e := range server.GetExamplesFromServer() {
				log.Printf("%s: %s\n", e.Name, e.ServiceName)
			}
		},
	}

	runExamplesCmd := &cobra.Command{
		Use:   "run-examples",
		Short: "Run all examples in the use command.",
		Long: `Run all examples generated by the use command through a
	provision/bind/unbind/deprovision cycle.

	Exits with a 0 if all examples were successful, 1 otherwise.`,
		Run: func(cmd *cobra.Command, args []string) {
			apiClient, err := client.NewClientFromEnv()
			if err != nil {
				log.Fatalf("Error creating client: %v", err)
			}

			switch {
			case exampleName != "" && serviceName == "":
				log.Fatalf("If an example name is specified, you must provide an accompanying service name.")
			case fileName != "":
				client.RunExamplesFromFile(apiClient, fileName, serviceName, exampleName)
			default:
				client.RunExamplesForService(server.GetExamplesFromServer(), apiClient, serviceName, exampleName, exampleJobCount)
			}
		},
	}

	clientCmd.AddCommand(clientCatalogCmd, provisionCmd, deprovisionCmd, bindCmd, unbindCmd, lastCmd, updateCmd, upgradeCmd)
	if featureflags.Enabled(featureflags.EnableLegacyExamplesCommands) {
		clientCmd.AddCommand(runExamplesCmd, examplesCmd)
	}

	bindFlag := func(dest *string, name, description string, commands ...*cobra.Command) {
		for _, sc := range commands {
			sc.Flags().StringVarP(dest, name, "", "", description)
			_ = sc.MarkFlagRequired(name)
		}
	}

	bindFlag(&instanceID, "instanceid", "id of the service instance to operate on (user defined)", provisionCmd, deprovisionCmd, bindCmd, unbindCmd, lastCmd, updateCmd, upgradeCmd)
	bindFlag(&serviceID, "serviceid", "GUID of the service instanceid references (see catalog)", provisionCmd, deprovisionCmd, bindCmd, unbindCmd, updateCmd, upgradeCmd)
	bindFlag(&planID, "planid", "GUID of the service instanceid references (see catalog entry for the associated serviceid)", provisionCmd, deprovisionCmd, bindCmd, unbindCmd, updateCmd, upgradeCmd)
	bindFlag(&bindingID, "bindingid", "GUID of the binding to work on (user defined)", bindCmd, unbindCmd)
	bindFlag(&oldVersion, "oldversion", "old terraform version", upgradeCmd)
	bindFlag(&newVersion, "newversion", "new terraform version", upgradeCmd)

	for _, sc := range []*cobra.Command{provisionCmd, bindCmd, updateCmd} {
		sc.Flags().StringVarP(&parametersJSON, "params", "", "{}", "JSON string of user-defined parameters to pass to the request")
	}

	runExamplesCmd.Flags().StringVarP(&serviceName, "service-name", "", "", "name of the service to run tests for")
	runExamplesCmd.Flags().StringVarP(&exampleName, "example-name", "", "", "only run examples matching this name")
	runExamplesCmd.Flags().StringVarP(&fileName, "filename", "", "", "json file that contains list of CompleteServiceExamples")
	runExamplesCmd.Flags().IntVarP(&exampleJobCount, "jobs", "j", 1, "number of parallel client examples to run concurrently")
}

func newClientCommand(use, short string, run func(*client.Client) *client.BrokerResponse) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			apiClient, err := client.NewClientFromEnv()
			if err != nil {
				log.Fatalf("Could not create API client: %s", err)
			}

			results := run(apiClient)
			utils.PrettyPrintOrExit(results)
		},
	}
}
