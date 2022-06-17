package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/command"
)

func main() {
	plugin.Start(new(command.UpgradePlugin))
}
