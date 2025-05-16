package credhubrepo_test

import (
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/credhubrepo"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/onsi/gomega/gstruct"
)

const (
	fakeUAAClientName   = "fake-uaa-client-name"
	fakeUAAClientSecret = "fake-uaa-client-secret"
	fakeUAAAccessToken  = "fake-uaa-access-token"
	fakeUAAIDToken      = "fake-uaa-id-token"
	fakeUAARefreshToken = "fake-uaa-refresh-token"
	fakeActor           = "mtls-app:fake-app-guid"
	fakeCredentialPath  = "/c/csb/my-lovely-service/fake-binding-id/secrets-and-services"
	fakeUUID            = "1fed7e7a-28ed-47ac-8b1b-ac35cc6f0406"
)

var _ = Describe("CredHub Repository", func() {
	var (
		fakeUAAServer     *ghttp.Server
		fakeCredHubServer *ghttp.Server
		repo              *credhubrepo.Repo
		skipTLSVerify     bool
		caCert            string
	)

	JustBeforeEach(func() {
		repo = must(credhubrepo.New(config.CredStoreConfig{
			CredHubURL:        localhost(fakeCredHubServer.URL()),
			UaaURL:            localhost(fakeUAAServer.URL()),
			UaaClientName:     fakeUAAClientName,
			UaaClientSecret:   fakeUAAClientSecret,
			SkipSSLValidation: skipTLSVerify,
			CACert:            caCert,
		}))
	})

	BeforeEach(func() {
		skipTLSVerify = false
		caCert = ""
	})

	AfterEach(func() {
		fakeUAAServer.Close()
		fakeCredHubServer.Close()
	})

	Describe("Save()", func() {
		BeforeEach(func() {
			fakeUAAServer = ghttp.NewServer()
			appendUAATokenHandler(fakeUAAServer)

			fakeCredHubServer = ghttp.NewServer()
			appendStoreHandlers(fakeCredHubServer)
		})

		It("performs the correct API calls", func() {
			By("calling the Save() method")
			ref, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
			Expect(err).NotTo(HaveOccurred())
			Expect(ref).To(SatisfyAny(HaveLen(1), HaveKeyWithValue("credhub-ref", fakeCredentialPath)))

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

		When("UAA auth fails", func() {
			BeforeEach(func() {
				fakeUAAServer.SetHandler(0, ghttp.RespondWith(http.StatusUnauthorized, `not allowed`))
			})

			It("returns an error", func() {
				ref, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).To(MatchError(`failed to store credential "/c/csb/my-lovely-service/fake-binding-id/secrets-and-services": unexpected status code 401, expecting 200, body: not allowed`))
				Expect(ref).To(BeNil())
			})
		})

		When("create credential request fails", func() {
			BeforeEach(func() {
				fakeCredHubServer.RouteToHandler(http.MethodPut, "/api/v1/data", ghttp.RespondWith(http.StatusBadRequest, `bad request`))
			})

			It("returns an error", func() {
				ref, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).To(MatchError(`failed to store credential "/c/csb/my-lovely-service/fake-binding-id/secrets-and-services": unexpected status code 400 for CredHub endpoint "/api/v1/data", expecting [200], body: bad request`))
				Expect(ref).To(BeNil())
			})
		})

		When("create permission request fails", func() {
			BeforeEach(func() {
				fakeCredHubServer.RouteToHandler(http.MethodPost, "/api/v2/permissions", ghttp.RespondWith(http.StatusBadRequest, `request bad`))
			})

			It("returns an error", func() {
				ref, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).To(MatchError(`failed to set permission on credential "/c/csb/my-lovely-service/fake-binding-id/secrets-and-services": unexpected status code 400 for CredHub endpoint "/api/v2/permissions", expecting [201], body: request bad`))
				Expect(ref).To(BeNil())
			})
		})
	})

	Describe("Delete()", func() {
		BeforeEach(func() {
			fakeUAAServer = ghttp.NewServer()
			appendUAATokenHandler(fakeUAAServer)

			fakeCredHubServer = ghttp.NewServer()
			fakeCredHubServer.RouteToHandler(http.MethodGet, "/api/v1/permissions", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/api/v1/permissions", fmt.Sprintf("credential_name=%s", fakeCredentialPath)),
				ghttp.VerifyContentType("application/json"),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.RespondWith(http.StatusOK, `{"credential_name":"/some-credential-name","permissions":[{"actor":"some-actor","path":"some-path","operations":["read"]}]}`),
			))
			fakeCredHubServer.RouteToHandler(http.MethodDelete, "/api/v1/data", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodDelete, "/api/v1/data", fmt.Sprintf("name=%s", fakeCredentialPath)),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.RespondWith(http.StatusNoContent, nil),
			))
			fakeCredHubServer.RouteToHandler(http.MethodGet, "/api/v2/permissions", ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodGet, "/api/v2/permissions", fmt.Sprintf("actor=some-actor&path=%s", fakeCredentialPath)),
				ghttp.VerifyContentType("application/json"),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{"path":"/some-path/*","operations":["read","write"],"actor":"some-actor","uuid":"%s"}`, fakeUUID)),
			))
			fakeCredHubServer.RouteToHandler(http.MethodDelete, fmt.Sprintf("/api/v2/permissions/%s", fakeUUID), ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodDelete, fmt.Sprintf("/api/v2/permissions/%s", fakeUUID)),
				ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
				ghttp.RespondWith(http.StatusOK, nil),
			))
		})

		It("performs the correct API calls", func() {
			By("calling the Delete() method")
			err := repo.Delete(fakeCredentialPath)
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

		When("UAA auth fails", func() {
			BeforeEach(func() {
				fakeUAAServer.SetHandler(0, ghttp.RespondWith(http.StatusUnauthorized, `not allowed`))
			})

			It("returns an error", func() {
				err := repo.Delete(fakeCredentialPath)
				Expect(err).To(MatchError(`failed to list permissions for credential "/c/csb/my-lovely-service/fake-binding-id/secrets-and-services": unexpected status code 401, expecting 200, body: not allowed`))
			})
		})

		When("list permissions request fails", func() {
			BeforeEach(func() {
				fakeCredHubServer.RouteToHandler(http.MethodGet, "/api/v1/permissions", ghttp.RespondWith(http.StatusBadRequest, `request bad`))
			})

			It("returns an error", func() {
				err := repo.Delete(fakeCredentialPath)
				Expect(err).To(MatchError(`failed to list permissions for credential "/c/csb/my-lovely-service/fake-binding-id/secrets-and-services": unexpected status code 400 for CredHub endpoint "/api/v1/permissions?credential_name=/c/csb/my-lovely-service/fake-binding-id/secrets-and-services", expecting [200], body: request bad`))
			})
		})

		When("get permission request fails", func() {
			BeforeEach(func() {
				fakeCredHubServer.RouteToHandler(http.MethodGet, "/api/v2/permissions", ghttp.RespondWith(http.StatusBadRequest, `request is bad`))
			})

			It("returns an error", func() {
				err := repo.Delete(fakeCredentialPath)
				Expect(err).To(MatchError(`failed to get permission "some-actor" for credential "/c/csb/my-lovely-service/fake-binding-id/secrets-and-services": unexpected status code 400 for CredHub endpoint "/api/v2/permissions?actor=some-actor&path=%2Fc%2Fcsb%2Fmy-lovely-service%2Ffake-binding-id%2Fsecrets-and-services", expecting [200], body: request is bad`))
			})
		})

		When("delete permission request fails", func() {
			BeforeEach(func() {
				fakeCredHubServer.RouteToHandler(http.MethodDelete, fmt.Sprintf("/api/v2/permissions/%s", fakeUUID), ghttp.RespondWith(http.StatusBadRequest, `bad req`))
			})

			It("returns an error", func() {
				err := repo.Delete(fakeCredentialPath)
				Expect(err).To(MatchError(`failed to delete permission ID "1fed7e7a-28ed-47ac-8b1b-ac35cc6f0406": unexpected status code 400 for CredHub endpoint "/api/v2/permissions/1fed7e7a-28ed-47ac-8b1b-ac35cc6f0406", expecting [200], body: bad req`))
			})
		})

		When("delete credential request fails", func() {
			BeforeEach(func() {
				fakeCredHubServer.RouteToHandler(http.MethodDelete, "/api/v1/data", ghttp.RespondWith(http.StatusBadRequest, `bad request`))
			})

			It("returns an error", func() {
				err := repo.Delete(fakeCredentialPath)
				Expect(err).To(MatchError(`failed to delete credential "/c/csb/my-lovely-service/fake-binding-id/secrets-and-services": unexpected status code 400 for CredHub endpoint "/api/v1/data?name=/c/csb/my-lovely-service/fake-binding-id/secrets-and-services", expecting [204], body: bad request`))
			})
		})
	})

	Describe("token expiry", func() {
		When("the token is valid", func() {
			BeforeEach(func() {
				fakeUAAServer = ghttp.NewServer()
				appendUAATokenHandler(fakeUAAServer)

				fakeCredHubServer = ghttp.NewServer()
				appendStoreHandlers(fakeCredHubServer)
			})

			It("does not reauthenticate", func() {
				By("calling the Repo() method multiple times")
				_, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).NotTo(HaveOccurred())
				_, err = repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).NotTo(HaveOccurred())
				_, err = repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).NotTo(HaveOccurred())

				By("checking that a UAA login was performed only once")
				Expect(fakeUAAServer.ReceivedRequests()).Should(HaveLen(1))
			})
		})

		When("the token has expired", func() {
			BeforeEach(func() {
				fakeUAAServer = ghttp.NewServer()
				fakeUAAServer.RouteToHandler(http.MethodPost, "/oauth/token", ghttp.CombineHandlers(
					ghttp.RespondWith(http.StatusOK, must(json.Marshal(map[string]any{
						"access_token":  fakeUAAAccessToken,
						"id_token":      fakeUAAIDToken,
						"refresh_token": fakeUAARefreshToken,
						"token_type":    "bearer",
						"expires_in":    1, // expires in 1 second
					}))),
				))

				fakeCredHubServer = ghttp.NewServer()
				appendStoreHandlers(fakeCredHubServer)
			})

			It("re-authenticates", func() {
				By("calling the Save() method once, triggering authentication")
				_, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUAAServer.ReceivedRequests()).Should(HaveLen(1))

				By("waiting for the token to expire")
				time.Sleep(1100 * time.Millisecond) // 1.1 seconds

				By("calling the Save() method again, triggering re-authentication")
				_, err = repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUAAServer.ReceivedRequests()).Should(HaveLen(2))
			})
		})

		When("the token has been invalidated", func() {
			const fakeOtherUAAAccessToken = "fake-other-uaa-access-token"

			BeforeEach(func() {
				fakeUAAServer = ghttp.NewServer()
				fakeUAAServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/oauth/token"),
						ghttp.RespondWith(http.StatusOK, must(json.Marshal(map[string]any{
							"access_token":  fakeUAAAccessToken,
							"id_token":      fakeUAAIDToken,
							"refresh_token": fakeUAARefreshToken,
							"token_type":    "bearer",
							"expires_in":    3600,
						}))),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/oauth/token"),
						ghttp.RespondWith(http.StatusOK, must(json.Marshal(map[string]any{
							"access_token":  fakeOtherUAAAccessToken, // New token
							"id_token":      fakeUAAIDToken,
							"refresh_token": fakeUAARefreshToken,
							"token_type":    "bearer",
							"expires_in":    3600,
						}))),
					),
				)

				fakeCredHubServer = ghttp.NewServer()
				fakeCredHubServer.AppendHandlers(
					// First call to Save()
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPut, "/api/v1/data", ""),
						ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
						ghttp.RespondWith(http.StatusOK, `{"type":"json","version_created_at":"2019-02-01T20:37:52Z","id":"25e00859-efc3-4a77-8822-2313ac127aa2","name":"/c/csb/my-lovely-service/fake-binding-id/secrets-and-services","metadata":{"description":"examplemetadata"},"value":{"foo":"bar"}}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/v2/permissions", ""),
						ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
						ghttp.RespondWith(http.StatusCreated, `{"path":"/some-path/*","operations":["read","write"],"actor":"some-actor","uuid":"d8863f92-f364-4a8e-afcc-d44bf1b453eb"}`),
					),

					// Second call to Save()
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPut, "/api/v1/data", ""),
						ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeUAAAccessToken),
						ghttp.RespondWith(http.StatusUnauthorized, `token no longer valid`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPut, "/api/v1/data", ""),
						ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeOtherUAAAccessToken),
						ghttp.RespondWith(http.StatusOK, `{"type":"json","version_created_at":"2019-02-01T20:37:52Z","id":"25e00859-efc3-4a77-8822-2313ac127aa2","name":"/c/csb/my-lovely-service/fake-binding-id/secrets-and-services","metadata":{"description":"examplemetadata"},"value":{"foo":"bar"}}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/v2/permissions", ""),
						ghttp.VerifyHeaderKV("Authorization", "Bearer "+fakeOtherUAAAccessToken),
						ghttp.RespondWith(http.StatusCreated, `{"path":"/some-path/*","operations":["read","write"],"actor":"some-actor","uuid":"d8863f92-f364-4a8e-afcc-d44bf1b453eb"}`),
					),
				)
			})

			It("re-authenticates", func() {
				By("calling Save() a first time which fetches and caches a token")
				_, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUAAServer.ReceivedRequests()).Should(HaveLen(1))
				Expect(fakeCredHubServer.ReceivedRequests()).Should(HaveLen(2))

				By("calling Save() a second time, which finds the token doesn't work and fetches another one")
				_, err = repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUAAServer.ReceivedRequests()).Should(HaveLen(2))
				Expect(fakeCredHubServer.ReceivedRequests()).Should(HaveLen(5))
			})
		})
	})

	Describe("TLS validation", func() {
		When("UAA server uses TLS with a certificate that can't be verified", func() {
			BeforeEach(func() {
				fakeUAAServer = ghttp.NewTLSServer() // Cert can't be verified
				appendUAATokenHandler(fakeUAAServer)

				fakeCredHubServer = ghttp.NewServer() // Not using TLS
				appendStoreHandlers(fakeCredHubServer)
			})

			It("fails with TLS validation errors", func() {
				_, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).To(MatchError(ContainSubstring("tls: failed to verify certificate: x509:")))
			})
		})

		When("CredHub server uses TLS with a certificate that can't be verified", func() {
			BeforeEach(func() {
				fakeUAAServer = ghttp.NewServer() // Not using TLS
				appendUAATokenHandler(fakeUAAServer)

				fakeCredHubServer = ghttp.NewTLSServer() // Cert can't be verified
				appendStoreHandlers(fakeCredHubServer)
			})

			It("fails with TLS validation errors", func() {
				_, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).To(MatchError(ContainSubstring("tls: failed to verify certificate: x509:")))
			})
		})

		When("both servers use TLS with a certificate that can't be verified", func() {
			BeforeEach(func() {
				fakeUAAServer = ghttp.NewTLSServer() // Cert can't be verified
				appendUAATokenHandler(fakeUAAServer)

				fakeCredHubServer = ghttp.NewTLSServer() // Cert can't be verified
				appendStoreHandlers(fakeCredHubServer)
			})

			It("fails with TLS validation errors", func() {
				_, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).To(MatchError(ContainSubstring("tls: failed to verify certificate: x509:")))
			})
		})

		When("TLS validation is skipped", func() {
			BeforeEach(func() {
				fakeUAAServer = ghttp.NewTLSServer() // Cert can't be verified
				appendUAATokenHandler(fakeUAAServer)

				fakeCredHubServer = ghttp.NewTLSServer() // Cert can't be verified
				appendStoreHandlers(fakeCredHubServer)

				skipTLSVerify = true
			})

			It("succeeds", func() {
				_, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).NotTo(HaveOccurred())

				// Check the servers did actually get the requests
				Expect(fakeUAAServer.ReceivedRequests()).Should(HaveLen(1))
				Expect(fakeCredHubServer.ReceivedRequests()).ShouldNot(BeEmpty())
			})
		})
	})

	Describe("CA certificate", func() {
		BeforeEach(func() {
			tlsConfig := tls.Config{
				Certificates: []tls.Certificate{must(tls.LoadX509KeyPair(filepath.Join(fakeCertsPath, "fakeLocalhost.crt"), filepath.Join(fakeCertsPath, "fakeLocalhost.key")))},
			}

			fakeUAAServer = ghttp.NewUnstartedServer()
			fakeUAAServer.HTTPTestServer.TLS = &tlsConfig
			fakeUAAServer.HTTPTestServer.StartTLS()
			appendUAATokenHandler(fakeUAAServer)

			fakeCredHubServer = ghttp.NewUnstartedServer()
			fakeCredHubServer.HTTPTestServer.TLS = &tlsConfig
			fakeCredHubServer.HTTPTestServer.StartTLS()
			appendStoreHandlers(fakeCredHubServer)
		})

		When("CA cert file is not provided", func() {
			It("fails to validate the server certificate", func() {
				_, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).To(MatchError(ContainSubstring("tls: failed to verify certificate: x509:")))
			})
		})

		When("CA cert file is provided", func() {
			BeforeEach(func() {
				caCert = string(must(os.ReadFile(filepath.Join(fakeCertsPath, "fakeRootCA.crt"))))
			})

			It("successfully validates the server certificate", func() {
				_, err := repo.Save(fakeCredentialPath, map[string]any{"foo": "bar"}, fakeActor)
				Expect(err).NotTo(HaveOccurred())

				// Check the servers did actually get the requests
				Expect(fakeUAAServer.ReceivedRequests()).Should(HaveLen(1))
				Expect(fakeCredHubServer.ReceivedRequests()).ShouldNot(BeEmpty())
			})
		})

		When("CA cert file is not valid", func() {
			It("fails to create the resource", func() {
				const invalidData = `not**valid  as CA**** cert`
				_, err := credhubrepo.New(config.CredStoreConfig{
					CACert: invalidData,
				})
				Expect(err).To(MatchError("failed to add CA cert to pool"))
			})
		})

	})
})

func appendUAATokenHandler(fakeUAAServer *ghttp.Server) {
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
}

func appendStoreHandlers(fakeCredHubServer *ghttp.Server) {
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
		ghttp.VerifyJSON(fmt.Sprintf(`{"actor":"mtls-app:fake-app-guid","operations":["read"],"path":"%s"}`, fakeCredentialPath)),
		ghttp.RespondWith(http.StatusCreated, `{"path":"/some-path/*","operations":["read","write"],"actor":"some-actor","uuid":"d8863f92-f364-4a8e-afcc-d44bf1b453eb"}`),
	))
}

// Replace IP address with host name so that certificate validation works
func localhost(url string) string {
	return strings.ReplaceAll(url, "127.0.0.1", "localhost")
}

func must[A any](input A, err error) A {
	GinkgoHelper()

	Expect(err).NotTo(HaveOccurred())
	return input
}
