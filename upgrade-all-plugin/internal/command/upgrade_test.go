package command_test

import (
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"

	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/command"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

type mi struct {
	Version string `json:"version"`
}

var _ = Describe("UpgradeAll", func() {

	var (
		fakeCliConnection *pluginfakes.FakeCliConnection
		fakeServer        *ghttp.Server
		responseCode      int
	)

	BeforeEach(func() {
		fakeCliConnection = &pluginfakes.FakeCliConnection{}
	})

	When("invalid input is given", func() {
		It("returns an error", func() {
			err := command.UpgradeAll(fakeCliConnection, []string{})

			Expect(err).To(MatchError(fmt.Errorf("broker name must be specifed")))
		})
	})

	When("access token can't be retrieved", func() {
		It("returns an error", func() {
			fakeCliConnection.ApiVersionReturns("3.0.0", nil)
			fakeCliConnection.IsLoggedInReturns(true, nil)
			fakeCliConnection.AccessTokenReturns("", fmt.Errorf("AccessToken error"))

			err := command.UpgradeAll(fakeCliConnection, []string{"broker-name"})
			Expect(err).To(MatchError(fmt.Errorf("error retrieving api access token: AccessToken error")))
		})
	})

	When("api endpoint can't be retrieved", func() {
		It("returns an error", func() {
			fakeCliConnection.ApiVersionReturns("3.0.0", nil)
			fakeCliConnection.IsLoggedInReturns(true, nil)
			fakeCliConnection.AccessTokenReturns("access-token", nil)
			fakeCliConnection.ApiEndpointReturns("", fmt.Errorf("APIEndpoint error"))

			err := command.UpgradeAll(fakeCliConnection, []string{"broker-name"})
			Expect(err).To(MatchError(fmt.Errorf("error retrieving api endpoint: APIEndpoint error")))
		})
	})

	Describe("success", func() {
		BeforeEach(func() {
			fakeCliConnection.ApiVersionReturns("3.0.0", nil)
			fakeCliConnection.IsLoggedInReturns(true, nil)
			fakeCliConnection.AccessTokenReturns("access-token", nil)
			responseCode = http.StatusOK
			responsePlans := ccapi.ServicePlans{
				Plans: []ccapi.Plan{
					{
						GUID: "test-guid",
						MaintenanceInfo: struct {
							Version string `json:"version"`
						}{Version: "test-version"},
					},
				},
			}
			responseServiceInstances := ccapi.ServiceInstances{Instances: []ccapi.ServiceInstance{{
				GUID:             "test-guid",
				UpgradeAvailable: true,
				Relationships: struct {
					ServicePlan struct {
						Data struct {
							GUID string `json:"guid"`
						} `json:"data"`
					} `json:"service_plan"`
				}{
					ServicePlan: struct {
						Data struct {
							GUID string `json:"guid"`
						} `json:"data"`
					}{
						Data: struct {
							GUID string `json:"guid"`
						}{
							GUID: "test-guid",
						},
					},
				},
				LastOperation: struct {
					Type        string `json:"type"`
					State       string `json:"state"`
					Description string `json:"description"`
				}{
					Type:        "",
					State:       "",
					Description: "",
				},
			}}}

			fakeServer = ghttp.NewServer()
			fakeServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v3/service_plans", "per_page=5000&service_broker_names=test-broker-name"),
					ghttp.VerifyHeaderKV("Authorization", "access-token"),
					ghttp.RespondWithJSONEncoded(responseCode, responsePlans),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v3/service_instances", "per_page=5000&service_plan_guids=test-guid"),
					ghttp.VerifyHeaderKV("Authorization", "access-token"),
					ghttp.RespondWithJSONEncoded(responseCode, responseServiceInstances),
				),
			)
			fakeCliConnection.ApiEndpointReturns(fakeServer.URL(), nil)

		})

		It("makes a request to get all service plans for a given broker", func() {
			err := command.UpgradeAll(fakeCliConnection, []string{"test-broker-name"})
			Expect(err).NotTo(HaveOccurred())
		})

		When("dryRun flag is given", func() {
			It("should return nil", func() {
				err := command.UpgradeAll(fakeCliConnection, []string{"test-broker-name", "-dry-run"})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		// Gets the service plans
		// Gets service instances
		// It upgrades the service instances with upgradeAvailable

	})

	//Describe("dryRun flag", func() {
	//	BeforeEach(func() {
	//		fakeCliConnection.ApiVersionReturns("3.0.0", nil)
	//		fakeCliConnection.IsLoggedInReturns(true, nil)
	//		fakeCliConnection.AccessTokenReturns("access-token", nil)
	//		fakeCliConnection.ApiEndpointReturns("api-endpoint", nil)
	//	})
	//
	//	When("flag provided ", func() {
	//		It("outputs the error", func() {
	//			fakeCliConnection.ApiVersionReturns("3.0.0", nil)
	//			fakeCliConnection.IsLoggedInReturns(false, nil)
	//
	//			err = command.UpgradeAll(&fakeCliConnection, []string{"broker-name"})
	//
	//			Expect(err).To(MatchError("you must authenticate with the cf cli before running this command"))
	//
	//		})
	//	})
	//	When("unable to check if logged in", func() {
	//		It("outputs the error", func() {
	//			fakeCliConnection.ApiVersionReturns("3.0.0", nil)
	//			fakeCliConnection.IsLoggedInReturns(false, fmt.Errorf("isLoggedIn error"))
	//
	//			err = command.UpgradeAll(&fakeCliConnection, []string{"broker-name"})
	//
	//			Expect(err).To(MatchError("error validating user authentication: isLoggedIn error"))
	//
	//		})
	//	})
	//
	//})

})
