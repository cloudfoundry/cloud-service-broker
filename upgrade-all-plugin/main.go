package main

import (
	"log"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/validate"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/command"

	"code.cloudfoundry.org/cli/plugin"
)

type UpgradePlugin struct{}

func (p *UpgradePlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "upgrade-all-service-instances" {
		l := log.New(os.Stdout, "", 0)
		err := command.UpgradeAll(cliConnection, args[1:], l)
		if err != nil {
			l.Printf("upgrade-all-service-instances plugin failed: %s", err.Error())
			os.Exit(1)
		}
	}
}

func (p *UpgradePlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "UpgradeAllServiceInstances",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 7,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "upgrade-all-service-instances",
				HelpText: "Upgrade all service instances from a broker to the latest available version of their current service plans.",
				UsageDetails: plugin.Usage{
					Usage: validate.Usage,
					Options: map[string]string{
						"-batch-size": "The number of concurrent upgrades (defaults to 10)",
					},
				},
			},
		},
	}
}

func main() {
	plugin.Start(&UpgradePlugin{})
}
