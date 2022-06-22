package ccapi_test

import (
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("GetServiceInstances", func() {

	var (
		fakeServer *ghttp.Server
		req        requester.Requester
		fakeCCAPI  ccapi.CCAPI
	)

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()
		DeferCleanup(fakeServer.Close)
		req = requester.NewRequester(fakeServer.URL(), "fake-token", false)
		fakeCCAPI = ccapi.NewCCAPI(req)
	})

	When("service instances exist in the given plans", func() {
		BeforeEach(func() {
			responseServiceInstances := `
			{
			  "resources": [{
				"guid": "test-guid",
				"maintenance_info": {
				  "version": "1.0.0"
				},
				"upgrade_available": false,
				"last_operation": {
				  "type": "create",
				  "state": "succeeded",
				  "description": "Operation succeeded",
				  "updated_at": "2020-03-10T15:49:32Z",
				  "created_at": "2020-03-10T15:49:29Z"
				},
				"relationships": {
				  "service_plan": {
					"data": {
					  "guid": "test-plan-guid"
					}
				  }
				}
			  }]
			}`
			fakeServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("GET", "/v3/service_instances", "per_page=5000&service_plan_guids=test-plan-guid,another-test-guid"),
					ghttp.RespondWith(http.StatusOK, responseServiceInstances),
				),
			)
		})

		It("returns instances from the given plans", func() {
			actualInstances, err := fakeCCAPI.GetServiceInstances([]string{"test-plan-guid", "another-test-guid"})

			By("checking the valid service instance is returned")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(actualInstances)).To(Equal(1))
			Expect(actualInstances[0].GUID).To(Equal("test-guid"))

			requests := fakeServer.ReceivedRequests()
			Expect(requests).To(HaveLen(1))

			By("making the appending the plan guids")
			Expect(requests[0].Method).To(Equal("GET"))
			Expect(requests[0].URL.Path).To(Equal("/v3/service_instances"))
			Expect(requests[0].URL.RawQuery).To(Equal("per_page=5000&service_plan_guids=test-plan-guid,another-test-guid"))
		})
	})

	When("no plan GUIDs are given", func() {
		It("returns an error", func() {
			actualInstances, err := fakeCCAPI.GetServiceInstances([]string{})

			Expect(err).To(MatchError("no service_plan_guids specified"))
			Expect(actualInstances).To(BeNil())
		})
	})

	When("the request fails", func() {
		BeforeEach(func() {

			fakeServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.RespondWith(http.StatusInternalServerError, nil),
				),
			)
		})

		It("returns an error", func() {
			_, err := fakeCCAPI.GetServiceInstances([]string{"test-guid"})

			Expect(err).To(MatchError("error getting service instances: http response: 500"))
		})
	})

})
