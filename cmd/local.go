package cmd

import (
	"log"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/local"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const localID = "local"

func init() {
	var params, plan, service, example string
	var all bool

	rootCmd.AddGroup(&cobra.Group{
		ID:    localID,
		Title: "Local Development",
	})

	marketplaceCmd := &cobra.Command{
		Use:     "marketplace",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: list services and plans",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			local.Marketplace(viper.GetString(pakCachePath))
		},
	}
	rootCmd.AddCommand(marketplaceCmd)

	const paramsFlag = "c"
	createServiceCmd := &cobra.Command{
		Use:     "create-service SERVICE PLAN NAME",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: create a service instance",
		Args:    cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			local.CreateService(args[0], args[1], args[2], params, viper.GetString(pakCachePath))
		},
	}
	createServiceCmd.Flags().StringVarP(&params, paramsFlag, paramsFlag, "", "parameters as JSON")
	rootCmd.AddCommand(createServiceCmd)

	servicesCmd := &cobra.Command{
		Use:     "services",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: list service instances",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			local.Services(viper.GetString(pakCachePath))
		},
	}
	rootCmd.AddCommand(servicesCmd)

	serviceCmd := &cobra.Command{
		Use:     "service",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: provides information on a service instance",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			local.Service(args[0], viper.GetString(pakCachePath))
		},
	}
	rootCmd.AddCommand(serviceCmd)

	const planFlag = "p"
	updateServiceCmd := &cobra.Command{
		Use:     "update-service NAME",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: update a service instance",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			local.UpdateService(args[0], plan, params, viper.GetString(pakCachePath))
		},
	}
	updateServiceCmd.Flags().StringVarP(&params, paramsFlag, paramsFlag, "", "parameters as JSON")
	updateServiceCmd.Flags().StringVarP(&plan, planFlag, planFlag, "", "change service plan for a service instance")
	rootCmd.AddCommand(updateServiceCmd)

	upgradeServiceCmd := &cobra.Command{
		Use:     "upgrade-service NAME PREVIOUS_VERSION",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: upgrade a service instance from PREVIOUS_VERSION to current version",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			local.UpgradeService(args[0], args[1], viper.GetString(pakCachePath))
		},
	}
	rootCmd.AddCommand(upgradeServiceCmd)

	deleteServiceCmd := &cobra.Command{
		Use:     "delete-service NAME",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: delete a service instance",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			local.DeleteService(args[0], viper.GetString(pakCachePath))
		},
	}
	rootCmd.AddCommand(deleteServiceCmd)

	createServiceKeyCmd := &cobra.Command{
		Use:     "create-service-key SERVICE_INSTANCE SERVICE_KEY",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: create a service key",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			local.CreateServiceKey(args[0], args[1], params, viper.GetString(pakCachePath))
		},
	}
	createServiceKeyCmd.Flags().StringVarP(&params, paramsFlag, paramsFlag, "", "parameters as JSON")
	rootCmd.AddCommand(createServiceKeyCmd)

	serviceKeyCmd := &cobra.Command{
		Use:     "service-key SERVICE_INSTANCE SERVICE_KEY",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: print a service key",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			local.ServiceKey(args[0], args[1])
		},
	}
	rootCmd.AddCommand(serviceKeyCmd)

	serviceKeysCmd := &cobra.Command{
		Use:     "service-keys NAME",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: list service keys for a service instance",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			local.ServiceKeys(args[0])
		},
	}
	rootCmd.AddCommand(serviceKeysCmd)

	deleteServiceKeyCmd := &cobra.Command{
		Use:     "delete-service-key SERVICE_INSTANCE SERVICE_KEY",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: delete a service key",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			local.DeleteServiceKey(args[0], args[1], viper.GetString(pakCachePath))
		},
	}
	rootCmd.AddCommand(deleteServiceKeyCmd)

	listExamplesCmd := &cobra.Command{
		Use:     "examples",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: list example tests",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			local.ListExamples(viper.GetString(pakCachePath))
		},
	}
	rootCmd.AddCommand(listExamplesCmd)

	const (
		serviceFlag = "service-name"
		exampleFlag = "example-name"
		allFlag     = "all"
	)
	runExamplesCmd := &cobra.Command{
		Use:     "run-examples",
		GroupID: localID,
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: run example tests",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if !all && service == "" && example == "" {
				log.Fatalln("specify --service-name and/or --example-name, or --all to run all the tests")
			}
			local.RunExamples(service, example, viper.GetString(pakCachePath))
		},
	}
	runExamplesCmd.Flags().StringVarP(&service, serviceFlag, "s", "", "service offering name")
	runExamplesCmd.Flags().StringVarP(&example, exampleFlag, "e", "", "example test name")
	runExamplesCmd.Flags().BoolVarP(&all, allFlag, "a", false, "run all tests")
	rootCmd.AddCommand(runExamplesCmd)
}
