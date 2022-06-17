package validate

import (
	"fmt"
	"regexp"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/hashicorp/go-version"
)

func ValidateInput(cliConnection plugin.CliConnection, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("broker name must be specifed")
	}

	err := sanitiseBrokerName(args[0])
	if err != nil {
		return err
	}

	err = validateAPIVersion(cliConnection)
	if err != nil {
		return err
	}

	err = validateLoggedIn(cliConnection)
	if err != nil {
		return err
	}

	return nil
}

func validateAPIVersion(cliConnection plugin.CliConnection) error {
	rawApiVersion, err := cliConnection.ApiVersion()
	if err != nil {
		return fmt.Errorf("error retrieving api version: %s", err)
	}
	apiVersion, err := version.NewVersion(rawApiVersion)

	if apiVersion.LessThan(version.Must(version.NewVersion("3.0.0"))) {
		return fmt.Errorf("plugin requires CF API version >= 3.0.0")
	}

	return nil
}

func validateLoggedIn(cliConnection plugin.CliConnection) error {
	isLoggedIn, err := cliConnection.IsLoggedIn()
	if err != nil {
		return fmt.Errorf("error validating user authentication: %s", err)
	}
	if !isLoggedIn {
		return fmt.Errorf("you must authenticate with the cf cli before running this command")
	}
	return nil
}

func sanitiseBrokerName(name string) error {
	if valid, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", name); !valid {
		return fmt.Errorf("invalid brokername format")
	}
	return nil
}
