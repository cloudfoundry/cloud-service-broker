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

var _ = Describe("GetServicePlans", func() {

	var (
		fakeServer    *ghttp.Server
		req           requester.Requester
		fakeRequester *requesterfakes.FakeRequester
	)

	When("Given a valid brokername", func() {
		BeforeEach(func() {
			responseServicePlans := ccapi.ServicePlans{Plans: []ccapi.Plan{{
				GUID: "test-guid",
				MaintenanceInfo: struct {
					Version string `json:"version"`
				}{
					Version: "test-version",
				},
			}}}

			fakeServer = ghttp.NewServer()
			fakeServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v3/service_plans", "per_page=5000&service_broker_names=test-broker-name"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, responseServicePlans),
				),
			)
			req = requester.NewRequester(fakeServer.URL(), "fake-token", false)

		})

		It("returns plans from that broker", func() {
			By("checking the brokername is in the query")
			actualPlans, err := ccapi.GetServicePlans(req, "test-broker-name")

			Expect(err).NotTo(HaveOccurred())

			By("checking the plan is returned")
			Expect(len(actualPlans)).To(Equal(1))
			Expect(actualPlans[0].GUID).To(Equal("test-guid"))
		})
	})

	When("the request fails", func() {
		BeforeEach(func() {
			fakeRequester = &requesterfakes.FakeRequester{}
			fakeRequester.GetReturns(fmt.Errorf("some error"))
		})

		It("returns an error", func() {
			_, err := ccapi.GetServicePlans(fakeRequester, "test-broker-name")

			Expect(err).To(MatchError("error getting service plans: some error"))
		})
	})

})
