package paramparser_test

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("ParseBindDetails", func() {
	It("can parse bind details", func() {
		fakeDomainBindDetails := domain.BindDetails{
			AppGUID:   "fake-app-guid",
			PlanID:    "fake-plan-id",
			ServiceID: "fake-service-id",
			BindResource: &domain.BindResource{
				AppGuid: "fake-bind-app-guid",
			},
			RawContext:    []byte(`{"foo": "bar"}`),
			RawParameters: []byte(`{"baz": "quz"}`),
		}

		bindDetails, err := paramparser.ParseBindDetails(fakeDomainBindDetails)

		Expect(err).NotTo(HaveOccurred())
		Expect(bindDetails).To(Equal(paramparser.BindDetails{
			AppGUID:        "fake-app-guid",
			PlanID:         "fake-plan-id",
			ServiceID:      "fake-service-id",
			BindAppGUID:    "fake-bind-app-guid",
			RequestParams:  map[string]any{"baz": "quz"},
			RequestContext: map[string]any{"foo": "bar"},
		}))
	})

	When("no bind_resource is instantiated", func() {
		It("succeeds", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
				AppGUID:   "fake-app-guid",
				PlanID:    "fake-plan-id",
				ServiceID: "fake-service-id",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(bindDetails).To(Equal(paramparser.BindDetails{
				AppGUID:   "fake-app-guid",
				PlanID:    "fake-plan-id",
				ServiceID: "fake-service-id",
			}))
		})
	})

	When("context is not valid JSON", func() {
		It("returns an error", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{BindResource: &domain.BindResource{}, RawContext: []byte("not-json")})

			Expect(err).To(MatchError(`error parsing request context: invalid character 'o' in literal null (expecting 'u')`))
			Expect(bindDetails).To(BeZero())
		})
	})

	When("params are not valid JSON", func() {
		It("returns an error", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{BindResource: &domain.BindResource{}, RawParameters: []byte("not-json")})

			Expect(err).To(MatchError(`error parsing request parameters: invalid character 'o' in literal null (expecting 'u')`))
			Expect(bindDetails).To(BeZero())
		})
	})

	When("input is empty", func() {
		It("succeeds with an empty result", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{BindResource: &domain.BindResource{}})

			Expect(err).NotTo(HaveOccurred())
			Expect(bindDetails).To(BeZero())
		})
	})
})
