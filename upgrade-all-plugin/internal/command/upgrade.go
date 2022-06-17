package command

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"time"

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

	NewRequester(apiEndPoint, accessToken, *skipVerify)

	return nil
}

type Requester struct {
	APIBaseURL string
	APIToken   string
	client     *http.Client
}

func NewRequester(apiBaseURL, apiToken string, insecureSkipVerify bool) Requester {
	return Requester{
		APIBaseURL: apiBaseURL,
		APIToken:   apiToken,
		client: &http.Client{
			Timeout: time.Minute,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
			},
		},
	}
}
