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

func (p *UpgradePlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{}
}
