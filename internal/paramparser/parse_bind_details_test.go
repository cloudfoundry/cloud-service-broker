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
				AppGuid:            "fake-bind-app-guid",
				SpaceGuid:          "fake-bind-space-guid",
				Route:              "fake-bind-route",
				CredentialClientID: "fake-bind-credential-client-id",
				BackupAgent:        false,
			},
			RawContext:    []byte(`{"foo": "bar"}`),
			RawParameters: []byte(`{"baz": "quz"}`),
		}

		bindDetails, err := paramparser.ParseBindDetails(fakeDomainBindDetails)

		Expect(err).NotTo(HaveOccurred())
		Expect(bindDetails).To(Equal(paramparser.BindDetails{
			AppGUID:                fakeDomainBindDetails.AppGUID,
			PlanID:                 fakeDomainBindDetails.PlanID,
			ServiceID:              fakeDomainBindDetails.ServiceID,
			BindAppGuid:            fakeDomainBindDetails.BindResource.AppGuid,
			BindSpaceGuid:          fakeDomainBindDetails.BindResource.SpaceGuid,
			BindRoute:              fakeDomainBindDetails.BindResource.Route,
			BindCredentialClientID: fakeDomainBindDetails.BindResource.CredentialClientID,
			BindBackupAgent:        false,
			RequestParams:          map[string]interface{}{"baz": "quz"},
			RequestContext:         map[string]interface{}{"foo": "bar"},
		}))
	})

	When("params are not valid JSON", func() {
		It("returns an error", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{BindResource: &domain.BindResource{}, RawParameters: []byte("not-json")})

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(`error parsing request parameters: invalid character 'o' in literal null (expecting 'u')`))
			Expect(bindDetails).To(BeZero())
		})
	})

	When("no bind_resource is instantiated", func() {
		It("returns an error", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{})

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(`error parsing bind request details: missing bind_resource`))
			Expect(bindDetails).To(BeZero())
		})
	})

	When("context is not valid JSON", func() {
		It("returns an error", func() {

		})
	})

	When("input is empty", func() {
		It("succeeds with an empty result", func() {

		})
	})
})
