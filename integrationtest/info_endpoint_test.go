package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	"github.com/cloudfoundry/cloud-service-broker/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Info Endpoint", func() {
	var (
		testHelper *helper.TestHelper
		session    *gexec.Session
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "info-endpoint")
		session = testHelper.StartBroker()
	})

	AfterEach(func() {
		session.Terminate()
	})

	It("responds to the info endpoint", func() {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/info", testHelper.Port))
		Expect(err).NotTo(HaveOccurred())
		Expect(resp).To(HaveHTTPStatus(http.StatusOK))

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		var data map[string]interface{}
		Expect(json.Unmarshal(body, &data)).NotTo(HaveOccurred())

		Expect(data).To(SatisfyAll(
			HaveKeyWithValue("version", utils.Version),
			HaveKeyWithValue("uptime", Not(BeEmpty())),
		))
	})
})
