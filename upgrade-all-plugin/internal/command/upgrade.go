package command

import (
	"flag"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/upgrader"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/validate"
)

func UpgradeAll(cliConnection plugin.CliConnection, args []string) error {
	err := validate.ValidateInput(cliConnection, args)
	if err != nil {
		return err
	}
	brokerName := args[0]

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
	batchSize := flagSet.Int("batch-size", 10, "number of concurrent upgrades")

	if len(args) > 1 {
		err = flagSet.Parse(args[1:])
		if err != nil {
			return err
		}
	}

	r := requester.NewRequester(apiEndPoint, accessToken, *skipVerify)
	api := ccapi.NewCCAPI(r)

	if err = upgrader.Upgrade(api, brokerName, *batchSize); err != nil {
		return err
	}

	return nil
}
