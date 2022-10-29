package integrationtest_test

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Terraform 0.12", func() {
	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	BeforeEach(func() {
		brokerpak = must(testdrive.BuildBrokerpak(csb, fixtures("terraform-0.12")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	It("can provision using this Terraform version", func() {
		const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
		_, err := broker.Provision(serviceOfferingGUID, servicePlanGUID)
		Expect(err).NotTo(HaveOccurred())
	})
})
