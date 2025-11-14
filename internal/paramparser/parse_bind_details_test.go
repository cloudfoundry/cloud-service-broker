package paramparser_test

import (
	"code.cloudfoundry.org/brokerapi/v13/domain"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/paramparser"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
)

var _ = Describe("ParseBindDetails", func() {
	It("can parse bind details", func() {
		fakeDomainBindDetails := domain.BindDetails{
			AppGUID:   "fake-app-guid-in-deprecated-location",
			PlanID:    "fake-plan-id",
			ServiceID: "fake-service-id",
			BindResource: &domain.BindResource{
				AppGuid:            "fake-app-guid",
				SpaceGuid:          "fake-space-guid",
				Route:              "fake-route",
				CredentialClientID: "fake-credential-client-id",
				BackupAgent:        true,
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
			RequestBindResource: map[string]any{
				"app_guid":             "fake-app-guid",
				"space_guid":           "fake-space-guid",
				"route":                "fake-route",
				"credential_client_id": "fake-credential-client-id",
				"backup_agent":         true,
			},
		}))
	})

	When("bind_resource is nil", func() {

		// for backwards combability app_guid is always stored at top-level and to bind_resource
		It("uses top-level app_guid for bind_resource", func() {
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
				RequestBindResource: map[string]any{
					"app_guid": "fake-app-guid",
				},
			}))
		})
	})

	When("bind_resource is empty", func() {

		// for backwards combability app_guid is always stored at top-level and to bind_resource
		It("uses top-level app_guid for bind_resource", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
				AppGUID:      "fake-app-guid",
				PlanID:       "fake-plan-id",
				ServiceID:    "fake-service-id",
				BindResource: &domain.BindResource{},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(bindDetails).To(Equal(paramparser.BindDetails{
				AppGUID:            "fake-app-guid",
				PlanID:             "fake-plan-id",
				ServiceID:          "fake-service-id",
				CredentialClientID: "",
				CredHubActor:       "mtls-app:fake-app-guid",
				RequestBindResource: map[string]any{
					"app_guid": "fake-app-guid",
				},
			}))
		})
	})

	When("app_guid is present in bind_resource and at the top level", func() {
		It("uses the app guid from bind_resource", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
				AppGUID:   "fake-app-guid-in-deprecated-location",
				PlanID:    "fake-plan-id",
				ServiceID: "fake-service-id",
				BindResource: &domain.BindResource{
					AppGuid: "fake-app-guid",
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(bindDetails).To(Equal(paramparser.BindDetails{
				AppGUID:            "fake-app-guid",
				PlanID:             "fake-plan-id",
				ServiceID:          "fake-service-id",
				CredentialClientID: "",
				CredHubActor:       "mtls-app:fake-app-guid",
				RequestBindResource: map[string]any{
					"app_guid": "fake-app-guid",
				},
			}))
		})
	})

	// Having the app guid at the top level is deprecated in favour of the app guid in bind_resource
	// In practice it's always present in both locations, but this could change in the future
	When("app guid is only present in bind_resource", func() {
		It("uses top-level app_guid for bind_resource", func() {
			bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
				BindResource: &domain.BindResource{
					AppGuid: "fake-app-guid",
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(bindDetails.AppGUID).To(Equal("fake-app-guid"))
			Expect(bindDetails.RequestBindResource).To(Equal(map[string]any{"app_guid": "fake-app-guid"}))
		})
	})

	// does not store bind_resource keys when keys are empty or go null values
	When("bind_resource has empty properties", func() {
		When("app_guid is empty", func() {
			It("uses the app guid from the top level", func() {
				bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
					AppGUID: "fake-app-guid",
					BindResource: &domain.BindResource{
						AppGuid: "",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(bindDetails.AppGUID).To(Equal("fake-app-guid"))
				Expect(bindDetails.RequestBindResource).To(Equal(map[string]any{"app_guid": "fake-app-guid"}))
			})
		})
		When("space_guid is empty", func() {
			It("omits space_guid from RequestBindResource", func() {
				bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
					BindResource: &domain.BindResource{
						AppGuid:   "fake-app-guid",
						SpaceGuid: "",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(bindDetails.AppGUID).To(Equal("fake-app-guid"))
				Expect(bindDetails.RequestBindResource).To(Equal(map[string]any{"app_guid": "fake-app-guid"}))
			})
		})
		When("route is empty", func() {
			It("omits route from RequestBindResource", func() {
				bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
					BindResource: &domain.BindResource{
						AppGuid: "fake-app-guid",
						Route:   "",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(bindDetails.AppGUID).To(Equal("fake-app-guid"))
				Expect(bindDetails.RequestBindResource).To(Equal(map[string]any{"app_guid": "fake-app-guid"}))
			})
		})
		When("credential_client_id is empty", func() {
			It("omits credential_client_id from RequestBindResource", func() {
				bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
					BindResource: &domain.BindResource{
						AppGuid:            "fake-app-guid",
						CredentialClientID: "",
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(bindDetails.AppGUID).To(Equal("fake-app-guid"))
				Expect(bindDetails.RequestBindResource).To(Equal(map[string]any{"app_guid": "fake-app-guid"}))
			})
		})
		When("backup_agent is false", func() {
			It("omits backup_agent from RequestBindResource", func() {
				bindDetails, err := paramparser.ParseBindDetails(domain.BindDetails{
					BindResource: &domain.BindResource{
						AppGuid:     "fake-app-guid",
						BackupAgent: false,
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(bindDetails.AppGUID).To(Equal("fake-app-guid"))
				Expect(bindDetails.RequestBindResource).To(Equal(map[string]any{"app_guid": "fake-app-guid"}))
			})
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
				Expect(err).To(MatchError(paramparser.ErrNoAppGUIDOrCredentialClient))
				Expect(bindDetails).To(BeZero())
			})
		})
	})
})

var _ = Describe("ParseStoredBindRequestDetails", func() {
	It("can parse stored bind request details", func() {

		serviceID := "fake-service-id"
		planID := "fake-plan-id"

		storedDetails := storage.BindRequestDetails{
			ServiceInstanceGUID: "fake-instance-guid",
			ServiceBindingGUID:  "fake-binding-guid",
			BindResource: storage.JSONObject{
				"app_guid":   "fake-app-guid",
				"space_guid": "fake-space-guid",
			},
			Parameters: storage.JSONObject{
				"foo": "bar",
			},
		}

		bindDetails, err := paramparser.ParseStoredBindRequestDetails(storedDetails, planID, serviceID)
		Expect(err).NotTo(HaveOccurred())
		Expect(bindDetails).To(Equal(paramparser.BindDetails{
			AppGUID:   "fake-app-guid",
			PlanID:    planID,
			ServiceID: serviceID,
			RequestParams: map[string]any{
				"foo": "bar",
			},
			RequestBindResource: map[string]any{
				"app_guid":   "fake-app-guid",
				"space_guid": "fake-space-guid",
			},
		}))
	})

	When("app_guid is missing from bind_resource", func() {
		serviceID := "fake-service-id"
		planID := "fake-plan-id"

		storedDetails := storage.BindRequestDetails{
			ServiceInstanceGUID: "fake-instance-guid",
			ServiceBindingGUID:  "fake-binding-guid",
			BindResource: storage.JSONObject{
				"space_guid": "fake-space-guid",
			},
			Parameters: storage.JSONObject{
				"foo": "bar",
			},
		}

		bindDetails, err := paramparser.ParseStoredBindRequestDetails(storedDetails, planID, serviceID)
		Expect(err).NotTo(HaveOccurred())
		Expect(bindDetails).To(Equal(paramparser.BindDetails{
			AppGUID:   "",
			PlanID:    planID,
			ServiceID: serviceID,
			RequestParams: map[string]any{
				"foo": "bar",
			},
			RequestBindResource: map[string]any{
				"space_guid": "fake-space-guid",
			},
		}))
	})
})
