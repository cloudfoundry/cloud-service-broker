package tf

import (
	"fmt"
	"regexp"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/executor"

	"code.cloudfoundry.org/lager/v3"
)

var planMessageMatcher = regexp.MustCompile(`Plan: \d+ to add, \d+ to change, (\d+) to destroy\.`)

func CheckTerraformPlanOutput(logger lager.Logger, output executor.ExecutionOutput) error {
	matches := planMessageMatcher.FindStringSubmatch(output.StdOut)
	switch {
	case len(matches) == 0: // presumably: "No changes. Infrastructure is up-to-date."
		logger.Info("no-match")
	case len(matches) == 2 && matches[1] == "0":
		logger.Info("no-destroyed")
	default:
		logger.Info("cancelling-destroy", lager.Data{"stdout": output.StdOut, "stderr": output.StdErr})
		return fmt.Errorf("tofu plan shows that resources would be destroyed - cancelling subsume")
	}
	return nil
}
