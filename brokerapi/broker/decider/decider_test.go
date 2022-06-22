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
			case "no MI in plan", "no params", "no MI change", "no plan":
			case "plan change":
				details.PlanID = "new-plan-id"
				details.PreviousPlanID = "previous-plan-id"
			case "plan static":
				details.PlanID = "static-plan-id"
				details.PreviousPlanID = "static-plan-id"
			case "no plan change":
				details.PlanID = "static-plan-id"
			case "params":
				details.RequestParams = map[string]any{"foo": "bar"}
			case "plan at v1":
				planVersion = version.Must(version.NewVersion("1.0.0"))
			case "MI none->v1":
				details.MaintenanceInfoVersion = version.Must(version.NewVersion("1.0.0"))
			case "MI none->v2":
				details.MaintenanceInfoVersion = version.Must(version.NewVersion("2.0.0"))
			case "MI v1->v1":
				details.MaintenanceInfoVersion = version.Must(version.NewVersion("1.0.0"))
				details.PreviousMaintenanceInfoVersion = version.Must(version.NewVersion("1.0.0"))
			case "MI v1->none":
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
	Entry(nil, "no MI in plan; no MI change; plan change;    no params", decider.Update, nil),
	Entry(nil, "no MI in plan; no MI change; no plan change; params", decider.Update, nil),
	Entry(nil, "no MI in plan; no MI change; plan static;    no params", decider.Update, nil),
	Entry(nil, "plan at v1;    no MI change; plan static;    no params", decider.Failed, "service instance needs to be upgraded before updating: maintenance info defined in broker service catalog, but not passed in request"),
	Entry(nil, "plan at v1;    MI v1->v1;    plan change;    no params", decider.Update, nil),
	Entry(nil, "plan at v1;    MI none->v1;  no plan change; params", decider.Update, nil),
	Entry(nil, "plan at v1;    MI v1->v1;    plan static;    no params", decider.Update, nil),
	Entry(nil, "plan at v1;    MI none->v1;  plan static;    no params", decider.Upgrade, nil),
	Entry(nil, "no MI in plan; MI v1->none;  plan static;    no params", decider.Upgrade, nil),
	Entry(nil, "plan at v1;    MI v2->v1;    plan static;    no params", decider.Upgrade, nil),
	Entry(nil, "plan at v1;    MI v1->v1;    plan change;    no params", decider.Update, nil), // Duplicate
	Entry(nil, "plan at v1;    MI v2->v1;    plan static;    params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "plan at v1;    MI v2->v1;    plan static;    params", decider.Failed, "service instance needs to be upgraded before updating"),
	Entry(nil, "plan at v1;    MI none->v2;  no plan change; no params", decider.Failed, apiresponses.ErrMaintenanceInfoConflict),
	Entry(nil, "no MI in plan; MI none->v1;  no plan change; no params", decider.Failed, apiresponses.ErrMaintenanceInfoNilConflict),
)
