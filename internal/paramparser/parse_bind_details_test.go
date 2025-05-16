package paramparser_test

import (
	"code.cloudfoundry.org/brokerapi/v13/domain"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/paramparser"
)

var _ = Describe("ParseBindDetails", func() {
	It("can parse bind details", func() {
		fakeDomainBindDetails := domain.BindDetails{
			AppGUID:   "fake-app-guid-in-deprecated-location",
			PlanID:    "fake-plan-id",
			ServiceID: "fake-service-id",
			BindResource: &domain.BindResource{
				AppGuid:            "fake-app-guid",
				CredentialClientID: "fake-credential-client-id",
			},
			RawContext:    []byte(`{"foo": "bar"}`),
			RawParameters: []byte(`{"baz": "quz"}`),
		}

		bindDetails, err := paramparser.ParseBindDetails(fakeDomainBindDetails)

		Expect(err).NotTo(HaveOccurred())
		Expect(bindDetails).To(Equal(paramparser.BindDetails{
			AppGUID:            "fake-app-guid",
			CredentialClientID: "fake-credential-client-id",
			PlanID:             "fake-plan-id",
			ServiceID:          "fake-service-id",
			CredHubActor:       "mtls-app:fake-app-guid",
			RequestParams:      map[string]any{"baz": "quz"},
			RequestContext:     map[string]any{"foo": "bar"},
		}))
	})

	When("no bind_resource is present", func() {
		It("succeeds", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
				AppGUID:   "fake-app-guid",
				PlanID:    "fake-plan-id",
				ServiceID: "fake-service-id",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(bindDetails).To(Equal(paramparser.BindDetails{
				AppGUID:            "fake-app-guid",
				PlanID:             "fake-plan-id",
				ServiceID:          "fake-service-id",
				CredentialClientID: "",
				CredHubActor:       "mtls-app:fake-app-guid",
			}))
		})
	})

	// Having the app guid at the top level is deprecated in favour of the app guid in bind_resource
	// In practice it's always present in both locations, but this could change in the future
	When("app guid is only present in bind_resource", func() {
		It("succeeds", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
				BindResource: &domain.BindResource{
					AppGuid: "fake-app-guid",
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(bindDetails.AppGUID).To(Equal("fake-app-guid"))
		})
	})

	When("bind_resource is present without an app guid", func() {
		It("uses the app guid from the top level", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
				AppGUID:      "fake-app-guid",
				BindResource: &domain.BindResource{}, // present, but empty
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(bindDetails.AppGUID).To(Equal("fake-app-guid"))
		})
	})

	When("context is not valid JSON", func() {
		It("returns an error", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{AppGUID: "fake-app-guid", RawContext: []byte("not-json")})

			Expect(err).To(MatchError(`error parsing request context: invalid character 'o' in literal null (expecting 'u')`))
			Expect(bindDetails).To(BeZero())
		})
	})

	When("params are not valid JSON", func() {
		It("returns an error", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{AppGUID: "fake-app-guid", RawParameters: []byte("not-json")})

			Expect(err).To(MatchError(`error parsing request parameters: invalid character 'o' in literal null (expecting 'u')`))
			Expect(bindDetails).To(BeZero())
		})
	})

	Describe("CredHubActor", func() {
		When("there's an app guid and client credential ID present", func() {
			It("uses the app guid", func() {
				bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
					AppGUID: "fake-app-guid",
					BindResource: &domain.BindResource{
						CredentialClientID: "fake-credential-client-id",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(bindDetails.CredHubActor).To(Equal("mtls-app:fake-app-guid"))
			})
		})

		When("there's a client credential ID but no app guid present", func() {
			It("uses the client credential ID", func() {
				bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
					BindResource: &domain.BindResource{
						CredentialClientID: "fake-credential-client-id",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(bindDetails.CredHubActor).To(Equal("uaa-client:fake-credential-client-id"))
			})
		})

		When("neither is present", func() {
			It("returns an error", func() {
				bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{})
				Expect(err).To(MatchError("no app GUID or credential client ID were provided in the binding request"))
				Expect(bindDetails).To(BeZero())
			})
		})
	})
})
