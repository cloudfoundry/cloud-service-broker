package integrationtest_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v2/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/testdrive"
)

var _ = Describe("Starting Server", func() {

	const userProvidedPlan = `[{"name": "user-plan-unique","id":"8b52a460-b246-11eb-a8f5-d349948e2481"}]`

	var brokerpak string

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("service-catalog")))

		DeferCleanup(func() {
			cleanup(brokerpak)
		})
	})

	When("TLS data is provided", func() {
		When("Valid data exists", func() {
			It("Should accept HTTPS requests", func() {
				isValid := true
				broker, err := testdrive.StartBroker(csb, brokerpak, database, testdrive.WithTLSConfig(isValid), testdrive.WithEnv(fmt.Sprintf("GSB_SERVICE_ALPHA_SERVICE_PLANS=%s", userProvidedPlan)), testdrive.WithOutputs(GinkgoWriter, GinkgoWriter))
				Expect(err).NotTo(HaveOccurred())

				_, err = http.Get(fmt.Sprintf("https://localhost:%d", broker.Port))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("Invalid data exists", func() {
			It("Should fail to start", func() {
				notValid := false
				_, err := testdrive.StartBroker(csb, brokerpak, database, testdrive.WithTLSConfig(notValid), testdrive.WithEnv(fmt.Sprintf("GSB_SERVICE_ALPHA_SERVICE_PLANS=%s", userProvidedPlan)), testdrive.WithOutputs(GinkgoWriter, GinkgoWriter))
				Expect(err).To(HaveOccurred())
			})
		})
	})

	When("No TLS data is provided", func() {
		It("Should return an error for HTTPS requests", func() {
			broker, err := testdrive.StartBroker(csb, brokerpak, database, testdrive.WithEnv(fmt.Sprintf("GSB_SERVICE_ALPHA_SERVICE_PLANS=%s", userProvidedPlan)), testdrive.WithOutputs(GinkgoWriter, GinkgoWriter))
			Expect(err).NotTo(HaveOccurred())

			_, err = http.Get(fmt.Sprintf("https://localhost:%d", broker.Port))
			Expect(err).To(HaveOccurred())
		})

		It("Should succeed for HTTP requests", func() {
			broker, err := testdrive.StartBroker(csb, brokerpak, database, testdrive.WithEnv(fmt.Sprintf("GSB_SERVICE_ALPHA_SERVICE_PLANS=%s", userProvidedPlan)), testdrive.WithOutputs(GinkgoWriter, GinkgoWriter))
			Expect(err).NotTo(HaveOccurred())

			_, err = http.Get(fmt.Sprintf("http://localhost:%d", broker.Port))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
