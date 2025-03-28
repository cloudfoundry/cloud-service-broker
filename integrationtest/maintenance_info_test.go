package integrationtest_test

import (
	"encoding/json"

	"code.cloudfoundry.org/brokerapi/v13/domain"
	"code.cloudfoundry.org/brokerapi/v13/domain/apiresponses"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v2/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/testdrive"
)

var _ = Describe("Maintenance Info", func() {
	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("maintenance-info")))
	})

	AfterEach(func() {
		Expect(broker.Stop()).To(Succeed())
		cleanup(brokerpak)
	})

	Context("Maintenance info", func() {
		When("TF upgrades are enabled", func() {
			BeforeEach(func() {
				broker = must(testdrive.StartBroker(
					csb, brokerpak, database,
					testdrive.WithOutputs(GinkgoWriter, GinkgoWriter),
					testdrive.WithTLSConfig(),
					testdrive.WithEnv("TERRAFORM_UPGRADES_ENABLED=true"),
				))
			})

			It("should match the default OpenTofu version", func() {
				catalogResponse := broker.Client.Catalog(requestID())
				Expect(catalogResponse.Error).NotTo(HaveOccurred())

				var catServices apiresponses.CatalogResponse
				Expect(json.Unmarshal(catalogResponse.ResponseBody, &catServices)).To(Succeed())
				Expect(catServices.Services[0].Plans[0].MaintenanceInfo).To(Equal(&domain.MaintenanceInfo{
					Public:      nil,
					Private:     "",
					Version:     "1.6.1",
					Description: "This upgrade provides support for OpenTofu version: 1.6.1. The upgrade operation will take a while. The instance and all associated bindings will be upgraded.",
				}))
			})
		})

		When("TF upgrades are disabled", func() {
			BeforeEach(func() {
				broker = must(testdrive.StartBroker(
					csb, brokerpak, database,
					testdrive.WithOutputs(GinkgoWriter, GinkgoWriter),
				))
			})

			It("should not be set for the plan", func() {
				catalogResponse := broker.Client.Catalog(requestID())

				Expect(catalogResponse.Error).NotTo(HaveOccurred())
				Expect(string(catalogResponse.ResponseBody)).ToNot(ContainSubstring(`"maintenance_info"`))
			})
		})
	})
})
