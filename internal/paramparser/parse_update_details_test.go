package paramparser_test

import (
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v9/domain"

	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"
)

var _ = Describe("ParseUpdateDetails", func() {
	It("can parse update details", func() {
		ud, err := paramparser.ParseUpdateDetails(domain.UpdateDetails{
			ServiceID: "fake-service-id",
			PlanID:    "fake-plan-id",
			MaintenanceInfo: &domain.MaintenanceInfo{
				Version: "1.2.3",
			},
			RawParameters: []byte(`{"foo":"bar"}`),
			PreviousValues: domain.PreviousValues{
				PlanID:    "fake-previous-plan-id",
				ServiceID: "fake-previous-service-id",
				OrgID:     "fake-previous-org-id",
				SpaceID:   "fake-previous-space-id",
				MaintenanceInfo: &domain.MaintenanceInfo{
					Version: "0.1.2",
				},
			},
			RawContext: []byte(`{"baz":"quz"}`),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(ud).To(Equal(paramparser.UpdateDetails{
			ServiceID:                      "fake-service-id",
			PlanID:                         "fake-plan-id",
			MaintenanceInfoVersion:         version.Must(version.NewVersion("1.2.3")),
			RequestParams:                  map[string]any{"foo": "bar"},
			RequestContext:                 map[string]any{"baz": "quz"},
			PreviousPlanID:                 "fake-previous-plan-id",
			PreviousServiceID:              "fake-previous-service-id",
			PreviousOrgID:                  "fake-previous-org-id",
			PreviousSpaceID:                "fake-previous-space-id",
			PreviousMaintenanceInfoVersion: version.Must(version.NewVersion("0.1.2")),
		}))
	})

	When("params are not valid JSON", func() {
		It("returns an error", func() {
			ud, err := paramparser.ParseUpdateDetails(domain.UpdateDetails{
				RawParameters: []byte(`not-json`),
			})

			Expect(err).To(MatchError(`error parsing request parameters: invalid character 'o' in literal null (expecting 'u')`))
			Expect(ud).To(BeZero())
		})
	})

	When("context is not valid JSON", func() {
		It("returns an error", func() {
			ud, err := paramparser.ParseUpdateDetails(domain.UpdateDetails{
				RawContext: []byte(`not-json`),
			})

			Expect(err).To(MatchError(`error parsing request context: invalid character 'o' in literal null (expecting 'u')`))
			Expect(ud).To(BeZero())
		})
	})

	When("maintenance info version is not valid", func() {
		It("returns an error", func() {
			ud, err := paramparser.ParseUpdateDetails(domain.UpdateDetails{
				MaintenanceInfo: &domain.MaintenanceInfo{
					Version: "not-a-valid-version",
				},
			})

			Expect(err).To(MatchError(`error parsing maintenance info: Malformed version: not-a-valid-version`))
			Expect(ud).To(BeZero())
		})
	})

	When("previous maintenance info version is not valid", func() {
		It("returns an error", func() {
			ud, err := paramparser.ParseUpdateDetails(domain.UpdateDetails{
				PreviousValues: domain.PreviousValues{
					MaintenanceInfo: &domain.MaintenanceInfo{
						Version: "not-a-valid-version",
					},
				},
			})

			Expect(err).To(MatchError(`error parsing previous maintenance info: Malformed version: not-a-valid-version`))
			Expect(ud).To(BeZero())
		})
	})

	When("input is empty", func() {
		It("succeeds with an empty result", func() {
			ud, err := paramparser.ParseUpdateDetails(domain.UpdateDetails{})

			Expect(err).NotTo(HaveOccurred())
			Expect(ud).To(BeZero())
		})
	})

	When("maintenance_info versions are empty", func() {
		It("succeeds with an empty result", func() {
			ud, err := paramparser.ParseUpdateDetails(domain.UpdateDetails{
				MaintenanceInfo: &domain.MaintenanceInfo{},
				PreviousValues: domain.PreviousValues{
					MaintenanceInfo: &domain.MaintenanceInfo{},
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(ud).To(BeZero())
		})
	})
})
