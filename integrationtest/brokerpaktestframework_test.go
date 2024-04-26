package integrationtest_test

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cloud-service-broker/v3/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("brokerpaktestframework", func() {
	It("works", func() {
		By("creating a mock Terraform")
		mockTerraform := must(brokerpaktestframework.NewTerraformMock(brokerpaktestframework.WithVersion("1.6.2")))

		By("building a fake brokerpak")
		cwd := must(os.Getwd())
		broker := must(brokerpaktestframework.BuildTestInstance(
			filepath.Join(cwd, "fixtures", "brokerpaktestframework"),
			mockTerraform,
			GinkgoWriter,
		))

		By("starting the broker")
		Expect(broker.Start(GinkgoWriter, nil)).To(Succeed())
		DeferCleanup(func() {
			Expect(broker.Cleanup()).To(Succeed())
		})

		By("testing catalog")
		catalog := must(broker.Catalog())
		Expect(catalog.Services).To(HaveLen(1))
		Expect(catalog.Services[0].Name).To(Equal("alpha-service"))

		By("testing provision")
		_, err := broker.Provision("does-not-exist", "does-not-exist", nil)
		Expect(err).To(MatchError(`cannot find service "does-not-exist" and plan "does-not-exist" in catalog`))
		serviceInstanceID := must(broker.Provision("alpha-service", "alpha", nil))
		Expect(serviceInstanceID).To(HaveLen(36))

		By("testing bind")
		_, err = broker.Bind("does-not-exist", "does-not-exist", "does-not-exist", nil)
		Expect(err).To(MatchError(`cannot find service "does-not-exist" and plan "does-not-exist" in catalog`))
		_, err = broker.Bind("alpha-service", "alpha", "does-not-exist", nil)
		Expect(err).To(MatchError(ContainSubstring(`error retrieving service instance details: could not find service instance details for: does-not-exist`)))
		creds := must(broker.Bind("alpha-service", "alpha", serviceInstanceID, nil))
		Expect(creds).To(BeEmpty())
	})
})
