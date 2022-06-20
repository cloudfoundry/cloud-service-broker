package command

import (
	"flag"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/validate"
)

func UpgradeAll(cliConnection plugin.CliConnection, args []string) error {
	err := validate.ValidateInput(cliConnection, args)
	if err != nil {
		return err
	}

	accessToken, err := cliConnection.AccessToken()
	if err != nil {
		return fmt.Errorf("error retrieving api access token: %s", err)
	}

	apiEndPoint, err := cliConnection.ApiEndpoint()
	if err != nil {
		return fmt.Errorf("error retrieving api endpoint: %s", err)
	}

	flagSet := flag.NewFlagSet("upgradeAll", flag.ExitOnError)
	//dryRun := flagSet.Bool("dry-run", false, "displays number of instances which would be upgraded")
	skipVerify := flagSet.Bool("skip-ssl-validation", false, "skip ssl certificate validation during http requests")

	if len(args) > 1 {
		err = flagSet.Parse(args[1:])
		if err != nil {
			return err
		}
	}

	requester.NewRequester(apiEndPoint, accessToken, *skipVerify)
	return nil
}
