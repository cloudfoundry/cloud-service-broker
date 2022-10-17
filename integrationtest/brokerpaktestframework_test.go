package integrationtest

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("brokerpaktestframework", func() {
	It("works", func() {
		By("creating a mock Terraform")
		mockTerraform, err := brokerpaktestframework.NewTerraformMock(brokerpaktestframework.WithVersion("1.2.3"))
		Expect(err).NotTo(HaveOccurred())

		By("building a fake brokerpak")
		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		broker, err := brokerpaktestframework.BuildTestInstance(
			filepath.Join(cwd, "fixtures", "brokerpaktestframework"),
			mockTerraform,
			GinkgoWriter,
		)
		Expect(err).NotTo(HaveOccurred())

		By("starting the broker")
		Expect(broker.Start(GinkgoWriter, nil)).To(Succeed())
		DeferCleanup(func() {
			Expect(broker.Cleanup()).To(Succeed())
		})

		By("testing catalog")
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())
		Expect(catalog.Services).To(HaveLen(1))
		Expect(catalog.Services[0].Name).To(Equal("alpha-service"))

		By("testing provision")
		_, err = broker.Provision("does-not-exist", "does-not-exist", nil)
		Expect(err).To(MatchError(`cannot find service "does-not-exist" and plan "does-not-exist" in catalog`))
		serviceInstanceID, err := broker.Provision("alpha-service", "alpha", nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(serviceInstanceID).To(HaveLen(36))

		By("testing bind")
		_, err = broker.Bind("does-not-exist", "does-not-exist", "does-not-exist", nil)
		Expect(err).To(MatchError(`cannot find service "does-not-exist" and plan "does-not-exist" in catalog`))
		_, err = broker.Bind("alpha-service", "alpha", "does-not-exist", nil)
		Expect(err).To(MatchError(ContainSubstring(`error retrieving service instance details: could not find service instance details for: does-not-exist`)))
		creds, err := broker.Bind("alpha-service", "alpha", serviceInstanceID, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(creds).To(BeEmpty())
	})
})
