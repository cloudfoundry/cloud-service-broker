package cmd

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/local"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	var (
		params string
		plan   string
	)

	marketplaceCmd := &cobra.Command{
		Use:   "marketplace",
		Short: "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: list services and plans",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			local.Marketplace(viper.GetString(pakCachePath))
		},
	}
	rootCmd.AddCommand(marketplaceCmd)

	const paramsFlag = "c"
	createServiceCmd := &cobra.Command{
		Use:   "create-service SERVICE PLAN NAME",
		Short: "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: create a service instance",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			local.CreateService(args[0], args[1], args[2], params, viper.GetString(pakCachePath))
		},
	}
	createServiceCmd.Flags().StringVarP(&params, paramsFlag, paramsFlag, "", "parameters as JSON")
	rootCmd.AddCommand(createServiceCmd)

	servicesCmd := &cobra.Command{
		Use:   "services",
		Short: "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: list service instances",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			local.Services(viper.GetString(pakCachePath))
		},
	}
	rootCmd.AddCommand(servicesCmd)

	const planFlag = "p"
	updateServiceCmd := &cobra.Command{
		Use:   "update-service NAME",
		Short: "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: update a service instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			local.UpdateService(args[0], plan, params, viper.GetString(pakCachePath))
		},
	}
	updateServiceCmd.Flags().StringVarP(&params, paramsFlag, paramsFlag, "", "parameters as JSON")
	updateServiceCmd.Flags().StringVarP(&plan, planFlag, planFlag, "", "change service plan for a service instance")
	rootCmd.AddCommand(updateServiceCmd)

	deleteServiceCmd := &cobra.Command{
		Use:   "delete-service NAME",
		Short: "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: delete a service instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			local.DeleteService(args[0], viper.GetString(pakCachePath))
		},
	}
	rootCmd.AddCommand(deleteServiceCmd)

	createServiceKeyCmd := &cobra.Command{
		Use:   "create-service-key SERVICE_INSTANCE SERVICE_KEY",
		Short: "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: create a service instance",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			local.CreateServiceKey(args[0], args[1], params, viper.GetString(pakCachePath))
		},
	}
	createServiceKeyCmd.Flags().StringVarP(&params, paramsFlag, paramsFlag, "", "parameters as JSON")
	rootCmd.AddCommand(createServiceKeyCmd)

	serviceKeysCmd := &cobra.Command{
		Use:   "service-keys NAME",
		Short: "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: list service keys for a service instance",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			local.ServiceKeys(args[0])
		},
	}
	rootCmd.AddCommand(serviceKeysCmd)

	deleteServiceKeyCmd := &cobra.Command{
		Use:   "delete-service-key SERVICE_INSTANCE SERVICE_KEY",
		Short: "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: delete a service key",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			local.DeleteServiceKey(args[0], args[1], viper.GetString(pakCachePath))
		},
	}
	rootCmd.AddCommand(deleteServiceKeyCmd)
}
