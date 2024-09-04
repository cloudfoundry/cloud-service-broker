package paramparser_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v11/domain"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/paramparser"
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
			RequestParams:    map[string]any{"foo": "bar"},
			RequestContext:   map[string]any{"baz": "quz"},
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

	When("a cloudfoundry context contains space_annotations or organization_annotations keys", func() {
		ctx := `{"platform": "cloudfoundry", "space_guid": "fake-space-guid", "space_annotations": {"a": "b"}, "organization_guid": "fake-org-guid", "organization_annotations": {"a": "b"}}`
		It("removes the annotation keys", func() {
			pd, err := paramparser.ParseProvisionDetails(domain.ProvisionDetails{
				RawContext: []byte(ctx),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(pd.RequestContext).ToNot(HaveKey("organization_annotations"))
			Expect(pd.RequestContext).ToNot(HaveKey("space_annotations"))
		})

		It("ignores all other keys and their values", func() {
			pd, err := paramparser.ParseProvisionDetails(domain.ProvisionDetails{
				RawContext: []byte(ctx),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(pd.RequestContext).To(HaveKey("platform"))
			Expect(pd.RequestContext["platform"]).To(Equal("cloudfoundry"))
			Expect(pd.RequestContext).To(HaveKey("space_guid"))
			Expect(pd.RequestContext["space_guid"]).To(Equal("fake-space-guid"))
			Expect(pd.RequestContext).To(HaveKey("organization_guid"))
			Expect(pd.RequestContext["organization_guid"]).To(Equal("fake-org-guid"))
		})
	})

	When("non-cloudfoundry context contains space_annotations or organization_annotations keys", func() {
		ctx := `{"platform": "kubernetes", "space_guid": "fake-space-guid", "space_annotations": {"a": "b"}, "organization_guid": "fake-org-guid", "organization_annotations": {"a": "b"}}`
		It("ignores the annotation keys and their values", func() {
			pd, err := paramparser.ParseProvisionDetails(domain.ProvisionDetails{
				RawContext: []byte(ctx),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(pd.RequestContext).To(HaveKey("organization_annotations"))
			Expect(pd.RequestContext["organization_annotations"]).To(HaveKey("a"))
			Expect(pd.RequestContext).To(HaveKey("space_annotations"))
			Expect(pd.RequestContext["space_annotations"]).To(HaveKey("a"))
		})
	})
})
