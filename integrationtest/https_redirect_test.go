package integrationtest_test

import (
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTPS redirect", func() {
	var broker *testdrive.Broker

	BeforeEach(func() {
		brokerpak := must(packer.BuildBrokerpak(csb, fixtures("https-redirect")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter), testdrive.WithHTTPRedirect()))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	It("redirects HTTP traffic", func() {
		By("redirecting an HTTP connection")
		errRedirected := fmt.Errorf("redirected")
		client := http.Client{
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return errRedirected
			},
		}

		_, err := client.Get(fmt.Sprintf("http://localhost:%d/info", broker.Port))
		Expect(err).To(MatchError(errRedirected))

		By("allowing a connection that originated as HTTPS")
		request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/info", broker.Port), nil)
		Expect(err).NotTo(HaveOccurred())
		request.Header.Add("X-Forwarded-Proto", "https")
		response, err := client.Do(request)
		Expect(err).NotTo(HaveOccurred())
		Expect(response).To(HaveHTTPStatus(http.StatusOK))
	})
})
