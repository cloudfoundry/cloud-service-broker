package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cloudfoundry/cloud-service-broker/internal/createservice"
)

func init() {
	var c string

	clientCmd := &cobra.Command{
		Use:     "create-service SERVICE PLAN NAME",
		Aliases: []string{"cs"},
		Short:   "EXPERIMENTAL AND SUBJECT TO BREAKING CHANGE: create a service instance",
		Args:    cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			createservice.Run(args[0], args[1], args[2], viper.GetString(pakCachePath))
		},
	}
	clientCmd.Flags().StringVarP(&c, "c", "c", "", "parameters as JSON")

	rootCmd.AddCommand(clientCmd)
}
