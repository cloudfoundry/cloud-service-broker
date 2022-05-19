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
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/dbservice"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/invoker"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func init() {
	var terraformProvider *tf.TerraformProvider
	var db *gorm.DB

	tfCmd := &cobra.Command{
		Use:   "tf",
		Short: "Interact with the Terraform backend",
		Long:  `Interact with the Terraform backend`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			logger := utils.NewLogger("tf")
			db = dbservice.New(logger)
			encryptor := setupDBEncryption(db, logger)
			store := storage.New(db, encryptor)
			terraformProvider = tf.NewTerraformProvider(
				executor.TFBinariesContext{},
				invoker.NewTerraformInvokerFactory(executor.NewExecutorFactory("", nil, nil), "", map[string]string{}),
				logger,
				tf.TfServiceDefinitionV1{},
				tf.NewDeploymentManager(store),
			)
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	rootCmd.AddCommand(tfCmd)

	dumpCmd := &cobra.Command{
		Use:   "dump",
		Short: "dump a Terraform workspace",
		Run: func(cmd *cobra.Command, args []string) {
			logger := utils.NewLogger("cloud-service-broker")
			encryptor := setupDBEncryption(db, logger)
			store := storage.New(db, encryptor)
			deployment, err := store.GetTerraformDeployment(args[0])
			if err != nil {
				log.Fatal(err)
			}
			ws := deployment.TFWorkspace()

			onlyState, err := cmd.Flags().GetBool("only-state")
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				log.Fatal(err)
			}

			if onlyState {
				fmt.Printf("%s", string(ws.State))
			} else {
				fmt.Println(ws)
			}
		},
	}
	dumpCmd.Flags().BoolP("only-state", "s", false, "dump the tf state file")
	tfCmd.AddCommand(dumpCmd)

	tfCmd.AddCommand(&cobra.Command{
		Use:   "wait",
		Short: "wait for a Terraform job",
		Run: func(cmd *cobra.Command, args []string) {
			err := terraformProvider.Wait(context.Background(), args[0])
			if err != nil {
				log.Fatal(err)
			}
		},
	})

	tfCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "show the list of Terraform workspaces",
		Run: func(cmd *cobra.Command, args []string) {
			results := []models.TerraformDeployment{}
			if err := db.Find(&results).Error; err != nil {
				log.Fatal(err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.StripEscape)
			fmt.Fprintln(w, "ID\tLast Operation\tState\tLast Updated\tElapsed\tMessage")

			for _, result := range results {
				lastUpdate := result.UpdatedAt.Format(time.RFC822)

				elapsed := ""
				if result.LastOperationState == tf.InProgress {
					elapsed = time.Since(result.UpdatedAt).Truncate(time.Second).String()
				}

				fmt.Fprintf(w, "%q\t%s\t%s\t%s\t%s\t%q\n", result.ID, result.LastOperationType, result.LastOperationState, lastUpdate, elapsed, result.LastOperationMessage)
			}
			w.Flush()
		},
	})
}
