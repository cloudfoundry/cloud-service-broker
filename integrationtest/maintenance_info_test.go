package integrationtest_test

import (
	"encoding/json"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)

var _ = Describe("Maintenance Info", func() {
	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "maintenance-info")
	})

	AfterEach(func() {
		session.Terminate().Wait()
	})

	Context("Maintenance info", func() {
		When("TF upgrades are enabled", func() {
			BeforeEach(func() {
				session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=true")
			})

			It("should match the default Terraform version", func() {
				catalogResponse := testHelper.Client().Catalog(requestID())
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
				session = testHelper.StartBroker()
			})

			It("should not be set for the plan", func() {
				catalogResponse := testHelper.Client().Catalog(requestID())

				Expect(catalogResponse.Error).NotTo(HaveOccurred())
				Expect(string(catalogResponse.ResponseBody)).ToNot(ContainSubstring(`"maintenance_info"`))
			})
		})

	})

})
