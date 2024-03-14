package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/cloudfoundry/cloud-service-broker/dbservice"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/utils"
)

func init() {
	purgeCmd := &cobra.Command{
		Use:     "purge-binding",
		GroupID: "broker",
		Short:   "purge a service binding from the database",
		Long: `Lets you remove a service binding (or service key) from the Cloud Service Broker database.

It does not actually delete the service binding, it just removes all references from the database.
This can be used to remove references to a service binding that has been manually removed,
or to clean up a service binding that fails to delete.

If using Cloud Foundry, identify the GUID of the service instance:

  cf service <name> --guid  # Prints the service instance guid

Then identify the GUID of the service binding, or service key that you want to remove.
You can see the service keys and bindings for a service instance by running:

  cf curl /v3/service_credential_bindings?service_instance_guids=<service-instance-guid>

Remove the binding from Cloud Service broker:

  cloud-service-broker purge <service-instance-guid> <service-binding-guid>

Then you can delete the binding from Cloud Foundry. Cloud Service Broker will confirm
to Cloud Foundry that the service binding or key no longer exists, and it will be removed
from the Cloud Foundry database

  cf unbind-service <app-name> <service-instance-name>

Or

  cf delete-service-key <service-instance-name> <service-key-name>
`,
		Run: func(cmd *cobra.Command, args []string) {
			switch len(args) {
			case 0:
				log.Fatal("missing service instance GUID and service binding GUID")
			case 1:
				log.Fatal("missing service binding GUID")
			case 2:
				purgeServiceBinding(args[0], args[1])
			default:
				log.Fatal("too many arguments")
			}
		},
	}

	rootCmd.AddCommand(purgeCmd)
}

func purgeServiceBinding(serviceInstanceGUID, serviceBindingGUID string) {
	logger := utils.NewLogger("purge-service-binding")
	db := dbservice.New(logger)
	encryptor := setupDBEncryption(db, logger)
	store := storage.New(db, encryptor)

	bindings, err := store.GetServiceBindingIDsForServiceInstance(serviceInstanceGUID)
	if err != nil {
		log.Fatalf("error listing bindings: %s", err)
	}
	for _, bindingGUID := range bindings {
		if bindingGUID == serviceBindingGUID {
			if err := deleteServiceBindingFromStore(store, serviceInstanceGUID, serviceBindingGUID); err != nil {
				log.Fatalf("error deleting binding %q for service instance %q: %s", serviceBindingGUID, serviceInstanceGUID, err)
			}
			log.Printf("deleted binding %q for service instance %q from the Cloud Service Broker database", serviceBindingGUID, serviceInstanceGUID)
			return
		}
	}

	log.Fatalf("could not find service binding %q for service instance %q", serviceBindingGUID, serviceInstanceGUID)
}

func deleteServiceBindingFromStore(store *storage.Storage, serviceInstanceGUID, serviceBindingGUID string) error {
	if err := store.DeleteServiceBindingCredentials(serviceBindingGUID, serviceInstanceGUID); err != nil {
		return fmt.Errorf("error deleting binding credentials for %q: %w", serviceBindingGUID, err)
	}
	if err := store.DeleteBindRequestDetails(serviceBindingGUID, serviceInstanceGUID); err != nil {
		return fmt.Errorf("error deleting binding request details for %q: %w", serviceBindingGUID, err)
	}
	if err := store.DeleteTerraformDeployment(fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, serviceBindingGUID)); err != nil {
		return fmt.Errorf("error deleting binding terraform deployment for %q: %s", serviceBindingGUID, err)
	}

	return nil
}
