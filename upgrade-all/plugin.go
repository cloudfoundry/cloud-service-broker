package main

import (
	"code.cloudfoundry.org/cli/plugin"
)

type UpgradePlugin struct{}

func (p *UpgradePlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if len(args) < 2 {
		panic("this should display plugin usage")
	}

	if args[0] == "upgrade-all-service-instances" {

		accessToken, err := cliConnection.AccessToken()
		if err != nil {
			panic(err)
		}

		apiEndPoint, err := cliConnection.ApiEndpoint()
		if err != nil {
			panic(err)
		}

		runUpgrade(accessToken, apiEndPoint, args[1])
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
					Usage: "upgrade-all-service-instances\n cf upgrade-all-service-instances",
				},
			},
		},
	}
}
