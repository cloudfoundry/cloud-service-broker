package command

import (
	"flag"
	"fmt"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/upgrader"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/validate"
)

func UpgradeAll(cliConnection plugin.CliConnection, args []string, log upgrader.Logger) error {
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

	sslSkipValidation, err := cliConnection.IsSSLDisabled()
	if err != nil {
		return fmt.Errorf("error retrieving api ssl validation status: %s", err)
	}

	flagSet := flag.NewFlagSet("upgradeAll", flag.ExitOnError)
	batchSize := flagSet.Int("batch-size", 10, "number of concurrent upgrades")
	if len(args) > 1 {
		err = flagSet.Parse(args[1:])
		if err != nil {
			return err
		}
	}

	r := requester.NewRequester(apiEndPoint, accessToken, sslSkipValidation)
	api := ccapi.NewCCAPI(r)

	if err := upgrader.Upgrade(api, brokerName, *batchSize, log); err != nil {
		return err
	}

	return nil
}
