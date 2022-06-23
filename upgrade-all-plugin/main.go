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
		Name: "UpgradeAllPlugin",
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
				HelpText: "all instances with an upgrade available will be upgraded.",
				UsageDetails: plugin.Usage{
					Usage: validate.Usage,
					Options: map[string]string{
						"-batch-size": "number of concurrent upgrades (default 10)",
					},
				},
			},
		},
	}
}

func main() {
	plugin.Start(&UpgradePlugin{})
}
