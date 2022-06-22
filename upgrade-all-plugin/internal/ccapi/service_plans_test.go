package ccapi_test

import (
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("GetServicePlans", func() {

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

	When("Given a valid brokername", func() {
		BeforeEach(func() {
			const response = `
			{
			  "resources": [
				{
				  "guid": "test-guid-1",
				  "maintenance_info": {
					"version": "test-mi-version"
				  }
				},
				{
				  "guid": "test-guid-2",
				  "maintenance_info": {
					"version": "test-mi-version"
				  }
				},
				{
				  "guid": "test-guid-3",
				  "maintenance_info": {
					"version": "test-mi-version"
				  }
				}
			  ]
			}
			`
			fakeServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("GET", "/v3/service_plans", "per_page=5000&service_broker_names=test-broker-name"),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)
		})

		It("returns plans from that broker", func() {
			By("checking the brokername is in the query")
			actualPlans, err := fakeCCAPI.GetServicePlans("test-broker-name")

			Expect(err).NotTo(HaveOccurred())

			By("checking the request contains the brokername query")
			requests := fakeServer.ReceivedRequests()
			Expect(requests).To(HaveLen(1))
			Expect(requests[0].Method).To(Equal("GET"))
			Expect(requests[0].URL.Path).To(Equal("/v3/service_plans"))
			Expect(requests[0].URL.RawQuery).To(Equal("per_page=5000&service_broker_names=test-broker-name"))

			By("checking the plan is returned")
			Expect(actualPlans).To(HaveLen(3))
			Expect(actualPlans[0].GUID).To(Equal("test-guid-1"))
			Expect(actualPlans[0].MaintenanceInfoVersion).To(Equal("test-mi-version"))
			Expect(actualPlans[1].GUID).To(Equal("test-guid-2"))
			Expect(actualPlans[1].MaintenanceInfoVersion).To(Equal("test-mi-version"))
			Expect(actualPlans[2].GUID).To(Equal("test-guid-3"))
			Expect(actualPlans[2].MaintenanceInfoVersion).To(Equal("test-mi-version"))
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
			_, err := fakeCCAPI.GetServicePlans("test-broker-name")

			Expect(err).To(MatchError("error getting service plans: http response: 500"))
		})
	})

})
