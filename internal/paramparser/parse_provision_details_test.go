package paramparser_test

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("ParseProvisionDetails", func() {
	It("can parse provision details", func() {
		pd, err := paramparser.ParseProvisionDetails(domain.ProvisionDetails{
			ServiceID:        "fake-service-id",
			PlanID:           "fake-plan-id",
			OrganizationGUID: "fake-org-guid",
			SpaceGUID:        "fake-space-guid",
			RawParameters:    []byte(`{"foo":"bar"}`),
			RawContext:       []byte(`{"baz":"quz"}`),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(pd).To(Equal(paramparser.ProvisionDetails{
			ServiceID:        "fake-service-id",
			PlanID:           "fake-plan-id",
			OrganizationGUID: "fake-org-guid",
			SpaceGUID:        "fake-space-guid",
			RequestParams:    map[string]interface{}{"foo": "bar"},
			RequestContext:   map[string]interface{}{"baz": "quz"},
		}))
	})

	When("params are not valid JSON", func() {
		It("returns an error", func() {
			pd, err := paramparser.ParseProvisionDetails(domain.ProvisionDetails{
				RawParameters: []byte(`not-json`),
			})

			Expect(err).To(MatchError(`error parsing request parameters: invalid character 'o' in literal null (expecting 'u')`))
			Expect(pd).To(BeZero())
		})
	})

	When("context is not valid JSON", func() {
		It("returns an error", func() {
			pd, err := paramparser.ParseProvisionDetails(domain.ProvisionDetails{
				RawContext: []byte(`not-json`),
			})

			Expect(err).To(MatchError(`error parsing request context: invalid character 'o' in literal null (expecting 'u')`))
			Expect(pd).To(BeZero())
		})
	})

	When("input is empty", func() {
		It("succeeds with an empty result", func() {
			pd, err := paramparser.ParseProvisionDetails(domain.ProvisionDetails{})

			Expect(err).NotTo(HaveOccurred())
			Expect(pd).To(BeZero())
		})
	})

	When("org guid is in context", func() {
		It("takes a string value from context", func() {
			pd, err := paramparser.ParseProvisionDetails(domain.ProvisionDetails{
				OrganizationGUID: "fake-org-guid-1",
				RawContext:       []byte(`{"organization_guid":"fake-org-guid-2"}`),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(pd.OrganizationGUID).To(Equal("fake-org-guid-2"))
		})

		It("ignores a non-string value from the context", func() {
			pd, err := paramparser.ParseProvisionDetails(domain.ProvisionDetails{
				OrganizationGUID: "fake-org-guid-1",
				RawContext:       []byte(`{"organization_guid":["fake-org-guid-2"]}`),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(pd.OrganizationGUID).To(Equal("fake-org-guid-1"))
		})
	})

	When("space guid is in context", func() {
		It("takes a string value from context", func() {
			pd, err := paramparser.ParseProvisionDetails(domain.ProvisionDetails{
				SpaceGUID:  "fake-space-guid-1",
				RawContext: []byte(`{"space_guid":"fake-space-guid-2"}`),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(pd.SpaceGUID).To(Equal("fake-space-guid-2"))
		})

		It("ignores a non-string value from the context", func() {
			pd, err := paramparser.ParseProvisionDetails(domain.ProvisionDetails{
				SpaceGUID:  "fake-space-guid-1",
				RawContext: []byte(`{"space_guid":["fake-space-guid-2"]}`),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(pd.SpaceGUID).To(Equal("fake-space-guid-1"))
		})
	})
})
