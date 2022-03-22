package paramparser_test

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("ParseUpdateDetails", func() {
	It("can parse update details", func() {
		ud, err := paramparser.ParseUpdateDetails(domain.UpdateDetails{
			ServiceID:     "fake-service-id",
			PlanID:        "fake-plan-id",
			RawParameters: []byte(`{"foo":"bar"}`),
			PreviousValues: domain.PreviousValues{
				PlanID:    "fake-previous-plan-id",
				ServiceID: "fake-previous-service-id",
				OrgID:     "fake-previous-org-id",
				SpaceID:   "fake-previous-space-id",
			},
			RawContext: []byte(`{"baz":"quz"}`),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(ud).To(Equal(paramparser.UpdateDetails{
			ServiceID:         "fake-service-id",
			PlanID:            "fake-plan-id",
			RequestParams:     map[string]interface{}{"foo": "bar"},
			RequestContext:    map[string]interface{}{"baz": "quz"},
			PreviousPlanID:    "fake-previous-plan-id",
			PreviousServiceID: "fake-previous-service-id",
			PreviousOrgID:     "fake-previous-org-id",
			PreviousSpaceID:   "fake-previous-space-id",
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

	When("input is empty", func() {
		It("succeeds with an empty result", func() {
			ud, err := paramparser.ParseUpdateDetails(domain.UpdateDetails{})

			Expect(err).NotTo(HaveOccurred())
			Expect(ud).To(BeZero())
		})
	})
})
