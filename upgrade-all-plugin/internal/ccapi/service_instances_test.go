package ccapi_test

import (
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester/requesterfakes"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("GetServiceInstances", func() {

	var (
		fakeServer    *ghttp.Server
		req           requester.Requester
		fakeRequester *requesterfakes.FakeRequester
	)

	When("Given a valid list of planGUIDs", func() {
		BeforeEach(func() {
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
					ghttp.VerifyRequest("GET", "/v3/service_instances", "per_page=5000&service_plan_guids=test-guid,another-test-guid"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, responseServiceInstances),
				),
			)
			req = requester.NewRequester(fakeServer.URL(), "fake-token", false)

		})

		It("returns instances from the given plans", func() {
			By("checking all plan GUIDs are appended to the query")
			actualInstances, err := ccapi.GetServiceInstances(req, []string{"test-guid", "another-test-guid"})

			Expect(err).NotTo(HaveOccurred())

			By("checking the valid service instance is returned")
			Expect(len(actualInstances)).To(Equal(1))
			Expect(actualInstances[0].GUID).To(Equal("test-guid"))
		})
	})

	When("no plan GUIDs are given", func() {
		It("returns an error", func() {
			actualInstances, err := ccapi.GetServiceInstances(req, []string{})

			Expect(err).To(MatchError("no service_plan_guids specified"))
			Expect(actualInstances).To(BeNil())
		})
	})

	When("the request fails", func() {
		BeforeEach(func() {
			fakeRequester = &requesterfakes.FakeRequester{}
			fakeRequester.GetReturns(fmt.Errorf("some error"))
		})

		It("returns an error", func() {
			_, err := ccapi.GetServiceInstances(fakeRequester, []string{"test-guid"})

			Expect(err).To(MatchError("error getting service instances: some error"))
		})
	})

})
