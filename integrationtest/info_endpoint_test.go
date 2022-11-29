package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	"github.com/cloudfoundry/cloud-service-broker/utils"
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
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/info", broker.Port))
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusOK))

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		var data map[string]any
		Expect(json.Unmarshal(body, &data)).NotTo(HaveOccurred())

		Expect(data).To(SatisfyAll(
			HaveKeyWithValue("version", utils.Version),
			HaveKeyWithValue("uptime", Not(BeEmpty())),
		))
	})
})
