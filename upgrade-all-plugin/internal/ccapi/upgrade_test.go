package ccapi_test

import (
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

// given an instance
// call patch on instance
// if error return error
// on success return nil

var _ = Describe("UpgradeServiceInstance", func() {

	var (
		fakeServer *ghttp.Server
		req        requester.Requester
	)

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()
		DeferCleanup(fakeServer.Close)
		req = requester.NewRequester(fakeServer.URL(), "fake-token", false)
	})

	When("given an upgradeable instance", func() {

		BeforeEach(func() {
			fakeServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", "/v3/service_instances/test-guid")))
		})

		It("successfully upgrades", func() {
			err := ccapi.UpgradeServiceInstance(req, "test-guid", "test-mi-version")
			Expect(err).NotTo(HaveOccurred())
			// want to assert that request was made

		})

	})

})
