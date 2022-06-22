package command

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/plugin"
)

const Usage = "cf upgrade-all-service-instances <broker-name>"

type UpgradePlugin struct{}

func (p *UpgradePlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "upgrade-all-service-instances" {
		err := UpgradeAll(cliConnection, args[1:])
		if err != nil {
			fmt.Printf("upgrade-all-service-instances plugin failed: %s", err.Error())
			os.Exit(1)
		}
	}

}

//TODO
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
					Usage: "upgrade-all-service-instances\n cf upgrade-all-service-instances",
				},
			},
		},
	}
}
