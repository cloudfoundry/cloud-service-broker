package cmd

import (
	"fmt"
	"log"

	"github.com/cloudfoundry/cloud-service-broker/dbservice"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	"github.com/spf13/cobra"
)

func init() {
	purgeCmd := &cobra.Command{
		Use:   "purge",
		Short: "purge a service instance from the database",
		Long: `Lets you remove a service instance from the Cloud Service Broker database.

It does not actually delete the service instance, it just removes all references from the database.
All bindings will also be removed from the database. This can be used to remove references to a service instance that
has been manually removed, or if you no longer want a service instance to be managed by the broker.

If using Cloud Foundry, the steps are:

  cf service <name> --guid  # Prints the service instance guid
  cloud-service-broker purge <guid>
  cf purge-service-instance <name>
`,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				log.Fatal("missing service instance GUID")
			case 1:
				purge(args[0])
			default:
				log.Fatal("too many arguments")
			}
		},
	}

	rootCmd.AddCommand(purgeCmd)
}

func purge(serviceInstanceGUID string) {
	logger := utils.NewLogger("purge")
	db := dbservice.New(logger)
	encryptor := setupDBEncryption(db, logger)
	store := storage.New(db, encryptor)

	bindings, err := store.GetServiceBindingsForServiceInstance(serviceInstanceGUID)
	if err != nil {
		log.Fatalf("error listing bindings: %s", err)
	}
	for _, bindingGUID := range bindings {
		if err := store.DeleteServiceBindingCredentials(bindingGUID, serviceInstanceGUID); err != nil {
			log.Fatalf("error deleting binding credentials for %q: %s", bindingGUID, err)
		}
		if err := store.DeleteBindRequestDetails(bindingGUID, serviceInstanceGUID); err != nil {
			log.Fatalf("error deleting binding request details for %q: %s", bindingGUID, err)
		}
		if err := store.DeleteTerraformDeployment(fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, bindingGUID)); err != nil {
			log.Fatalf("error deleting binding terraform deployment for %q: %s", bindingGUID, err)
		}
	}
	if err := store.DeleteProvisionRequestDetails(serviceInstanceGUID); err != nil {
		log.Fatalf("error deleting provision request details for %q: %s", serviceInstanceGUID, err)
	}
	if err := store.DeleteServiceInstanceDetails(serviceInstanceGUID); err != nil {
		log.Fatalf("error deleting service instance details for %q: %s", serviceInstanceGUID, err)
	}
	if err := store.DeleteTerraformDeployment(fmt.Sprintf("tf:%s:", serviceInstanceGUID)); err != nil {
		log.Fatalf("error deleting service terraform deployment for %q: %s", serviceInstanceGUID, err)
	}
}
