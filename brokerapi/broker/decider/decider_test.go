package decider_test

import (
	"fmt"
	"regexp"

	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/decider"
	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
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
			case "no MI in plan", "no params", "no request MI":
			case "plan change":
				details.PlanID = "new-plan-id"
				details.PreviousPlanID = "previous-plan-id"
			case "plan unchanged":
				details.PreviousPlanID = "plan-id"
			case "plan at v1":
				planVersion = version.Must(version.NewVersion("1.0.0"))
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
	Entry(nil, "no MI in plan; no request MI; plan unchanged; no params", decider.Update, nil),
	Entry(nil, "no MI in plan; no request MI; plan change;    no params", decider.Update, nil),
	Entry(nil, "no MI in plan; no request MI; plan unchanged; params", decider.Update, nil),
	Entry(nil, "no MI in plan; no request MI; plan change;    params", decider.Update, nil),

	// New world Updates, MI is in previous values and is not being changed
	Entry(nil, "plan at v1; MI unchanged; plan change;    no params", decider.Update, nil),
	Entry(nil, "plan at v1; MI unchanged; plan unchanged; params", decider.Update, nil),
	Entry(nil, "plan at v1; MI unchanged; plan change;    params", decider.Update, nil),

	// Adding, removing and changing MI
	Entry(nil, "plan at v1;    MI none->v1;  plan unchanged; no params", decider.Upgrade, nil),
	Entry(nil, "no MI in plan; MI v1->none;  plan unchanged; no params", decider.Upgrade, nil),
	Entry(nil, "plan at v1;    MI v2->v1;    plan unchanged; no params", decider.Upgrade, nil),

	// Combined Upgrade and Update
	Entry(nil, "plan at v1;    MI v2->v1;   plan unchanged; params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "plan at v1;    MI v2->v1;   plan change;    no params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "plan at v1;    MI v2->v1;   plan change;    params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "no MI in plan; MI v1->none; plan change;    no params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "no MI in plan; MI v1->none; plan unchanged; params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "no MI in plan; MI v1->none; plan change;    params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "plan at v1;    MI none->v1; plan change;    no params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "plan at v1;    MI none->v1; plan unchanged; params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "plan at v1;    MI none->v1; plan change;    params", decider.Failed, "service instance needs to be upgraded before updating"),

	// Attempted upgrades where the requested MI does not match the plan MI
	Entry(nil, "plan at v1;    MI none->v2; plan unchanged; no params", decider.Failed, apiresponses.ErrMaintenanceInfoConflict),
	Entry(nil, "no MI in plan; MI none->v1; plan unchanged; no params", decider.Failed, apiresponses.ErrMaintenanceInfoNilConflict),

	// Edge case: updates where MI is held constant
	// With CloudFoundry, new MI is only specified if it's different to previous MI, so we do not see this.
	Entry(nil, "plan at v1; MI v1->v1; plan unchanged; no params", decider.Update, nil),
	Entry(nil, "plan at v1; MI v1->v1; plan change;    no params", decider.Update, nil),
	Entry(nil, "plan at v1; MI v1->v1; plan unchanged; params", decider.Update, nil),
	Entry(nil, "plan at v1; MI v1->v1; plan change;    params", decider.Update, nil),

	// Edge case: updates where MI is not specified at all in the request as it has not changed.
	// With CloudFoundry, previous MI is always specified, so we do not see this.
	Entry(nil, "plan at v1; no request MI; plan unchanged; no params", decider.Update, nil),
	Entry(nil, "plan at v1; no request MI; plan change;    no params", decider.Update, nil),
	Entry(nil, "plan at v1; no request MI; plan unchanged; params", decider.Update, nil),
	Entry(nil, "plan at v1; no request MI; plan change;    params", decider.Update, nil),
)
