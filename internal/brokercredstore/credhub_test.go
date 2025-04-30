package brokercredstore_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"code.cloudfoundry.org/lager/v3/lagertest"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/brokercredstore"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/config"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/credstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/onsi/gomega/gstruct"
)

var _ = Describe("CredHub API", func() {
	const (
		fakeUAAClientName   = "fake-uaa-client-name"
		fakeUAAClientSecret = "fake-uaa-client-secret"
		fakeUAAAccessToken  = "fake-uaa-access-token"
		fakeUAAIDToken      = "fake-uaa-id-token"
		fakeUAARefreshToken = "fake-uaa-refresh-token"
		fakeServiceName     = "my-lovely-service"
		fakeBindingID       = "fake-binding-id"
		fakeAppGUID         = "fake-app-guid"
		fakeCredentialName  = "/c/csb/my-lovely-service/fake-binding-id/secrets-and-services"
		fakeUUID            = "1fed7e7a-28ed-47ac-8b1b-ac35cc6f0406"
	)

	var (
		fakeUAAServer     *ghttp.Server
		fakeCredHubServer *ghttp.Server
		fakeLogger        *lagertest.TestLogger
		credHubStore      brokercredstore.BrokerCredstore
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
			}))),
		))

		fakeCredHubServer = ghttp.NewServer()

		fakeLogger = lagertest.NewTestLogger("credhub-test")

		csConfig := config.CredStoreConfig{
			CredHubURL:        fakeCredHubServer.URL(),
			UaaURL:            fakeUAAServer.URL(),
			UaaClientName:     fakeUAAClientName,
			UaaClientSecret:   fakeUAAClientSecret,
			SkipSSLValidation: false,
			CACert:            "",
		}
		cs, err := credstore.NewCredhubStore(&csConfig, fakeLogger)
		Expect(err).NotTo(HaveOccurred())
		credHubStore = brokercredstore.NewBrokerCredstore(cs)
	})

	AfterEach(func() {
		fakeUAAServer.Close()
		fakeCredHubServer.Close()
	})

	Describe("Store()", func() {
		BeforeEach(func() {
			fakeCredHubServer.RouteToHandler(http.MethodGet, "/info", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/info", ""),
				ghttp.VerifyContentType("application/json"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]any{
					"auth-server": map[string]any{"url": fakeUAAServer.URL()},
					"app":         map[string]any{"name": "CredHub"},
				}),
			))
			fakeCredHubServer.RouteToHandler(http.MethodGet, "/version", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/version", ""),
				ghttp.VerifyContentType("application/json"),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.RespondWith(http.StatusOK, `{"version":"2.13.0"}`),
			))
			fakeCredHubServer.RouteToHandler(http.MethodPut, "/api/v1/data", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPut, "/api/v1/data", ""),
				ghttp.VerifyContentType("application/json"),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.VerifyJSON(`{"name":"/c/csb/my-lovely-service/fake-binding-id/secrets-and-services","type":"json","value":{"foo":"bar"}}`),
				ghttp.RespondWith(http.StatusOK, `{"type" : "json","version_created_at" : "2019-02-01T20:37:52Z","id":"25e00859-efc3-4a77-8822-2313ac127aa2","name":"/c/csb/my-lovely-service/fake-binding-id/secrets-and-services","metadata":{"description":"example metadata"},"value":{"foo":"bar"}}`),
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
			ref, err := credHubStore.Store(map[string]any{"foo": "bar"}, fakeServiceName, fakeBindingID, fakeAppGUID)
			Expect(err).NotTo(HaveOccurred())
			Expect(ref).To(Equal(map[string]any{"credhub-ref": fakeCredentialName}))

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
			fakeCredHubServer.RouteToHandler(http.MethodGet, "/info", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/info", ""),
				ghttp.VerifyContentType("application/json"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]any{
					"auth-server": map[string]any{"url": fakeUAAServer.URL()},
					"app":         map[string]any{"name": "CredHub"},
				}),
			))
			fakeCredHubServer.RouteToHandler(http.MethodGet, "/version", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/version", ""),
				ghttp.VerifyContentType("application/json"),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.RespondWith(http.StatusOK, `{"version":"2.13.0"}`),
			))
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
			credHubStore.Delete(fakeLogger, fakeServiceName, fakeBindingID)

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
})

func must[A any](input A, err error) A {
	GinkgoHelper()

	Expect(err).NotTo(HaveOccurred())
	return input
}
