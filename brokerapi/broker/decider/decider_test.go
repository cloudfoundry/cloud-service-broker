package decider_test

import (
	"fmt"
	"regexp"

	"code.cloudfoundry.org/brokerapi/v13/domain/apiresponses"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v2/brokerapi/broker/decider"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/paramparser"
)

var separator = regexp.MustCompile(`\s*;\s*`)

var _ = DescribeTable(
	"DecideOperation()",
	func(spec string, expectedOperation decider.Operation, expectedError any) {
		var (
			details     paramparser.UpdateDetails
			planVersion *version.Version
		)
		for _, token := range separator.Split(spec, -1) {
			switch token {
			case "service has no MI", "no params", "no request MI":
			case "plan change":
				details.PlanID = "new-plan-id"
				details.PreviousPlanID = "previous-plan-id"
			case "plan unchanged":
				details.PreviousPlanID = "plan-id"
			case "service at v1":
				planVersion = version.Must(version.NewVersion("1.0.0"))
			case "service at v2":
				planVersion = version.Must(version.NewVersion("2.0.0"))
			case "params":
				details.RequestParams = map[string]any{"foo": "bar"}
			case "MI none->v1":
				details.MaintenanceInfoVersion = version.Must(version.NewVersion("1.0.0"))
			case "MI none->v2":
				details.MaintenanceInfoVersion = version.Must(version.NewVersion("2.0.0"))
			case "MI v1->v1":
				details.MaintenanceInfoVersion = version.Must(version.NewVersion("1.0.0"))
				details.PreviousMaintenanceInfoVersion = version.Must(version.NewVersion("1.0.0"))
			case "MI v1->none", "MI unchanged":
				details.PreviousMaintenanceInfoVersion = version.Must(version.NewVersion("1.0.0"))
			case "MI v2->v1":
				details.MaintenanceInfoVersion = version.Must(version.NewVersion("1.0.0"))
				details.PreviousMaintenanceInfoVersion = version.Must(version.NewVersion("2.0.0"))
			default:
				Fail(fmt.Sprintf("invalid token: %s", token))
			}
		}

		operation, err := decider.DecideOperation(planVersion, details)
		switch expectedError {
		case nil:
			Expect(err).NotTo(HaveOccurred())
		default:
			Expect(err).To(MatchError(expectedError))
		}
		Expect(operation).To(Equal(expectedOperation))
	},
	// Old world Updates - MI does not exist
	Entry(nil, "service has no MI; no request MI; plan unchanged; no params", decider.Update, nil),
	Entry(nil, "service has no MI; no request MI; plan change;    no params", decider.Update, nil),
	Entry(nil, "service has no MI; no request MI; plan unchanged; params", decider.Update, nil),
	Entry(nil, "service has no MI; no request MI; plan change;    params", decider.Update, nil),

	// New world Updates, MI is in previous values and is not being changed
	Entry(nil, "service at v1; MI unchanged; plan change;    no params", decider.Update, nil),
	Entry(nil, "service at v1; MI unchanged; plan unchanged; params", decider.Update, nil),
	Entry(nil, "service at v1; MI unchanged; plan change;    params", decider.Update, nil),

	// Adding, removing and changing MI
	Entry(nil, "service at v1;     MI none->v1;  plan unchanged; no params", decider.Upgrade, nil),
	Entry(nil, "service has no MI; MI v1->none;  plan unchanged; no params", decider.Upgrade, nil),
	Entry(nil, "service at v1;     MI v2->v1;    plan unchanged; no params", decider.Upgrade, nil),

	// Combined Upgrade and Update
	Entry(nil, "service at v1;     MI v2->v1;   plan unchanged; params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "service at v1;     MI v2->v1;   plan change;    no params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "service at v1;     MI v2->v1;   plan change;    params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "service has no MI; MI v1->none; plan change;    no params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "service has no MI; MI v1->none; plan unchanged; params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "service has no MI; MI v1->none; plan change;    params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "service at v1;     MI none->v1; plan change;    no params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "service at v1;     MI none->v1; plan unchanged; params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "service at v1;     MI none->v1; plan change;    params", decider.Failed, "service instance needs to be upgraded before updating"),

	// Attempted upgrades where the requested MI does not match the plan MI
	Entry(nil, "service at v1;     MI none->v2; plan unchanged; no params", decider.Failed, apiresponses.ErrMaintenanceInfoConflict),
	Entry(nil, "service has no MI; MI none->v1; plan unchanged; no params", decider.Failed, apiresponses.ErrMaintenanceInfoNilConflict),

	// Updates where MI is held constant
	// In this case the Upgrade would be a no-op, so we default to it being a valid Update
	Entry(nil, "service at v1; MI v1->v1; plan unchanged; no params", decider.Update, nil),
	Entry(nil, "service at v1; MI v1->v1; plan change;    no params", decider.Update, nil),
	Entry(nil, "service at v1; MI v1->v1; plan unchanged; params", decider.Update, nil),
	Entry(nil, "service at v1; MI v1->v1; plan change;    params", decider.Update, nil),

	// When previous MI does not match the plan, and it's not an upgrade, it suggests the platform
	// (typically CloudFoundry) is not in sync with the broker. With CloudFoundry the "cf update-service-broker"
	// command needs to be run
	Entry(nil, "service at v2; MI unchanged;  plan unchanged; no params", decider.Failed, apiresponses.ErrMaintenanceInfoConflict),
	Entry(nil, "service at v2; MI unchanged;  plan change;    no params", decider.Failed, apiresponses.ErrMaintenanceInfoConflict),
	Entry(nil, "service at v2; MI unchanged;  plan unchanged; params", decider.Failed, apiresponses.ErrMaintenanceInfoConflict),
	Entry(nil, "service at v2; MI unchanged;  plan change;    params", decider.Failed, apiresponses.ErrMaintenanceInfoConflict),
	Entry(nil, "service at v1; no request MI; plan unchanged; no params", decider.Failed, apiresponses.ErrMaintenanceInfoConflict),
	Entry(nil, "service at v1; no request MI; plan change;    no params", decider.Failed, apiresponses.ErrMaintenanceInfoConflict),
	Entry(nil, "service at v1; no request MI; plan unchanged; params", decider.Failed, apiresponses.ErrMaintenanceInfoConflict),
	Entry(nil, "service at v1; no request MI; plan change;    params", decider.Failed, apiresponses.ErrMaintenanceInfoConflict),
)
