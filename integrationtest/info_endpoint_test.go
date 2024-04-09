package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/v2/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/testdrive"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Info Endpoint", func() {
	var broker *testdrive.Broker

	BeforeEach(func() {
		brokerpak := must(packer.BuildBrokerpak(csb, fixtures("info-endpoint")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	It("responds to the info endpoint", func() {
		resp := must(http.Get(fmt.Sprintf("http://localhost:%d/info", broker.Port)))
		Expect(resp).To(HaveHTTPStatus(http.StatusOK))

		defer resp.Body.Close()
		body := must(io.ReadAll(resp.Body))
		var data map[string]any
		Expect(json.Unmarshal(body, &data)).NotTo(HaveOccurred())

		Expect(data).To(SatisfyAll(
			HaveKeyWithValue("version", utils.Version),
			HaveKeyWithValue("uptime", Not(BeEmpty())),
		))
	})
})
