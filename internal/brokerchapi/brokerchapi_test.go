package brokerchapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/brokerchapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/onsi/gomega/gstruct"
)

var _ = Describe("Broker CredHub API", func() {
	const (
		fakeUAAClientName   = "fake-uaa-client-name"
		fakeUAAClientSecret = "fake-uaa-client-secret"
		fakeUAAAccessToken  = "fake-uaa-access-token"
		fakeUAAIDToken      = "fake-uaa-id-token"
		fakeUAARefreshToken = "fake-uaa-refresh-token"
		fakeAppGUID         = "fake-app-guid"
		fakeCredentialName  = "/c/csb/my-lovely-service/fake-binding-id/secrets-and-services"
		fakeUUID            = "1fed7e7a-28ed-47ac-8b1b-ac35cc6f0406"
	)

	var (
		fakeUAAServer     *ghttp.Server
		fakeCredHubServer *ghttp.Server
		store             *brokerchapi.Store
	)

	BeforeEach(func() {
		fakeUAAServer = ghttp.NewServer()
		fakeUAAServer.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest(http.MethodPost, "/oauth/token"),
			ghttp.VerifyHeader(http.Header{
				"Accept":       []string{"application/json"},
				"Content-Type": []string{"application/x-www-form-urlencoded"},
			}),
			ghttp.VerifyForm(url.Values{
				"client_id":     []string{fakeUAAClientName},
				"client_secret": []string{fakeUAAClientSecret},
				"grant_type":    []string{"client_credentials"},
				"response_type": []string{"token"},
			}),
			ghttp.RespondWith(http.StatusOK, must(json.Marshal(map[string]any{
				"access_token":  fakeUAAAccessToken,
				"id_token":      fakeUAAIDToken,
				"refresh_token": fakeUAARefreshToken,
				"token_type":    "bearer",
				"expires_in":    3600,
			}))),
		))

		fakeCredHubServer = ghttp.NewServer()

		store = must(brokerchapi.New(brokerchapi.Config{
			CredHubURL:      fakeCredHubServer.URL(),
			UAAURL:          fakeUAAServer.URL(),
			UAAClientName:   fakeUAAClientName,
			UAAClientSecret: fakeUAAClientSecret,
		}))
	})

	AfterEach(func() {
		fakeUAAServer.Close()
		fakeCredHubServer.Close()
	})

	Describe("Store()", func() {
		BeforeEach(func() {
			fakeCredHubServer.RouteToHandler(http.MethodPut, "/api/v1/data", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPut, "/api/v1/data", ""),
				ghttp.VerifyContentType("application/json"),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.VerifyJSON(`{"name":"/c/csb/my-lovely-service/fake-binding-id/secrets-and-services","type":"json","value":{"foo":"bar"}}`),
				ghttp.RespondWith(http.StatusOK, `{"type":"json","version_created_at":"2019-02-01T20:37:52Z","id":"25e00859-efc3-4a77-8822-2313ac127aa2","name":"/c/csb/my-lovely-service/fake-binding-id/secrets-and-services","metadata":{"description":"examplemetadata"},"value":{"foo":"bar"}}`),
			))
			fakeCredHubServer.RouteToHandler(http.MethodPost, "/api/v2/permissions", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPost, "/api/v2/permissions", ""),
				ghttp.VerifyContentType("application/json"),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.VerifyJSON(fmt.Sprintf(`{"actor":"mtls-app:fake-app-guid","operations":["read"],"path":"%s"}`, fakeCredentialName)),
				ghttp.RespondWith(http.StatusCreated, `{"path":"/some-path/*","operations":["read","write"],"actor":"some-actor","uuid":"d8863f92-f364-4a8e-afcc-d44bf1b453eb"}`),
			))
		})

		It("performs the correct API calls", func() {
			By("calling the Store() method")
			err := store.Save(map[string]any{"foo": "bar"}, fakeCredentialName, fakeAppGUID)
			Expect(err).NotTo(HaveOccurred())

			By("checking that a UAA login was performed")
			Expect(fakeUAAServer.ReceivedRequests()).Should(HaveLen(1))

			By("checking that the correct CredHub endpoints were called")
			Expect(fakeCredHubServer.ReceivedRequests()).To(ContainElements(
				gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Method": Equal(http.MethodPut),
					"URL": gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Path": Equal("/api/v1/data"),
					})),
				})),
				gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Method": Equal(http.MethodPost),
					"URL": gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Path": Equal("/api/v2/permissions"),
					})),
				})),
			))
		})
	})

	Describe("Delete()", func() {
		BeforeEach(func() {
			fakeCredHubServer.RouteToHandler(http.MethodGet, "/api/v1/permissions", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/api/v1/permissions", fmt.Sprintf("credential_name=%s", fakeCredentialName)),
				ghttp.VerifyContentType("application/json"),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.RespondWith(http.StatusOK, `{"credential_name":"/some-credential-name","permissions":[{"actor":"some-actor","path":"some-path","operations":["read"]}]}`),
			))
			fakeCredHubServer.RouteToHandler(http.MethodDelete, "/api/v1/data", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodDelete, "/api/v1/data", fmt.Sprintf("name=%s", fakeCredentialName)),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.RespondWith(http.StatusNoContent, nil),
			))
			fakeCredHubServer.RouteToHandler(http.MethodGet, "/api/v2/permissions", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/api/v2/permissions", fmt.Sprintf("actor=some-actor&path=%s", fakeCredentialName)),
				ghttp.VerifyContentType("application/json"),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{"path":"/some-path/*","operations":["read","write"],"actor":"some-actor","uuid":"%s"}`, fakeUUID)),
			))
			fakeCredHubServer.RouteToHandler(http.MethodDelete, fmt.Sprintf("/api/v2/permissions/%s", fakeUUID), ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodDelete, fmt.Sprintf("/api/v2/permissions/%s", fakeUUID)),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.RespondWith(http.StatusNoContent, nil),
			))
		})

		It("performs the correct API calls", func() {
			By("calling the Delete() method")
			err := store.Delete(fakeCredentialName)
			Expect(err).NotTo(HaveOccurred())

			By("checking that a UAA login was performed")
			Expect(fakeUAAServer.ReceivedRequests()).Should(HaveLen(1))

			By("checking that the correct CredHub endpoints were called")
			Expect(fakeCredHubServer.ReceivedRequests()).To(ContainElements(
				gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Method": Equal(http.MethodGet),
					"URL": gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Path": Equal("/api/v1/permissions"),
					})),
				})),
				gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Method": Equal(http.MethodDelete),
					"URL": gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Path": Equal("/api/v1/data"),
					})),
				})),
				gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Method": Equal(http.MethodGet),
					"URL": gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Path": Equal("/api/v2/permissions"),
					})),
				})),
				gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Method": Equal(http.MethodDelete),
					"URL": gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Path": Equal(fmt.Sprintf("/api/v2/permissions/%s", fakeUUID)),
					})),
				})),
			))
		})
	})

	Describe("token expiry", func() {
		BeforeEach(func() {
			fakeCredHubServer.RouteToHandler(http.MethodPut, "/api/v1/data", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPut, "/api/v1/data", ""),
				ghttp.VerifyContentType("application/json"),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.RespondWith(http.StatusOK, `{"type" : "json","version_created_at" : "2019-02-01T20:37:52Z","id":"25e00859-efc3-4a77-8822-2313ac127aa2","name":"/c/csb/my-lovely-service/fake-binding-id/secrets-and-services","metadata":{"description":"example metadata"},"value":{"foo":"bar"}}`),
			))
			fakeCredHubServer.RouteToHandler(http.MethodPost, "/api/v2/permissions", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPost, "/api/v2/permissions", ""),
				ghttp.VerifyContentType("application/json"),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.RespondWith(http.StatusCreated, `{"path":"/some-path/*","operations":["read","write"],"actor":"some-actor","uuid":"d8863f92-f364-4a8e-afcc-d44bf1b453eb"}`),
			))
		})

		When("the token is valid", func() {
			It("does not reauthenticate", func() {
				By("calling the Store() method multiple times")
				err := store.Save(map[string]any{"foo": "bar"}, fakeCredentialName, fakeAppGUID)
				Expect(err).NotTo(HaveOccurred())
				err = store.Save(map[string]any{"bar": "baz"}, fakeCredentialName, fakeAppGUID)
				Expect(err).NotTo(HaveOccurred())
				err = store.Save(map[string]any{"baz": "quz"}, fakeCredentialName, fakeAppGUID)
				Expect(err).NotTo(HaveOccurred())

				By("checking that a UAA login was performed only once")
				Expect(fakeUAAServer.ReceivedRequests()).Should(HaveLen(1))
			})
		})

		When("the token has expired", func() {
			BeforeEach(func() {
				fakeUAAServer.Reset()
				fakeUAAServer.RouteToHandler(http.MethodPost, "/oauth/token", ghttp.CombineHandlers(
					ghttp.RespondWith(http.StatusOK, must(json.Marshal(map[string]any{
						"access_token":  fakeUAAAccessToken,
						"id_token":      fakeUAAIDToken,
						"refresh_token": fakeUAARefreshToken,
						"token_type":    "bearer",
						"expires_in":    -10, // already expired
					}))),
				))
			})

			It("re-authenticates", func() {
				By("calling the Store() method once, triggering authentication")
				err := store.Save(map[string]any{"foo": "bar"}, fakeCredentialName, fakeAppGUID)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUAAServer.ReceivedRequests()).Should(HaveLen(1))

				By("calling the Store() method again, triggering re-authentication")
				err = store.Save(map[string]any{"bar": "baz"}, fakeCredentialName, fakeAppGUID)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUAAServer.ReceivedRequests()).Should(HaveLen(2))
			})
		})
	})
})

func must[A any](input A, err error) A {
	GinkgoHelper()

	Expect(err).NotTo(HaveOccurred())
	return input
}
