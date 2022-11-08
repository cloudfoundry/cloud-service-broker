package integrationtest_test

import (
	"encoding/json"

	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)

var _ = Describe("Maintenance Info", func() {
	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	BeforeEach(func() {
		brokerpak = must(testdrive.BuildBrokerpak(csb, fixtures("maintenance-info")))
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
					testdrive.WithEnv("TERRAFORM_UPGRADES_ENABLED=true"),
				))
			})

			It("should match the default Terraform version", func() {
				catalogResponse := broker.Client.Catalog(requestID())
				Expect(catalogResponse.Error).NotTo(HaveOccurred())

				var catServices apiresponses.CatalogResponse
				err := json.Unmarshal(catalogResponse.ResponseBody, &catServices)
				Expect(err).NotTo(HaveOccurred())
				Expect(catServices.Services[0].Plans[0].MaintenanceInfo).To(Equal(&domain.MaintenanceInfo{
					Public:      nil,
					Private:     "",
					Version:     "0.13.7",
					Description: "This upgrade provides support for Terraform version: 0.13.7. The upgrade operation will take a while. The instance and all associated bindings will be upgraded.",
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
