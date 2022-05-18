package integrationtest_test

import (
	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("MaintenanceInfo", func() {
	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "brokerpak-terraform-mi")
	})

	AfterEach(func() {
		session.Terminate()
		testHelper.Restore()
	})

	Context("Maintenance info", func() {
		When("TF upgrades are enabled", func() {
			BeforeEach(func() {
				session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=true")
			})

			It("should match the default Terraform version", func() {
				catalogResponse := testHelper.Client().Catalog(requestID())

				Expect(catalogResponse.Error).NotTo(HaveOccurred())
				Expect(string(catalogResponse.ResponseBody)).To(ContainSubstring(`"maintenance_info":{"version":"0.13.7","description":"This upgrade provides support for Terraform version: 0.13.7. The upgrade operation will take a while and all instances and bindings will be updated."}`))

			})
		})

		When("TF upgrades are disabled", func() {
			BeforeEach(func() {
				session = testHelper.StartBroker()
			})

			It("should not be set for the plan", func() {
				catalogResponse := testHelper.Client().Catalog(requestID())

				Expect(catalogResponse.Error).NotTo(HaveOccurred())
				Expect(string(catalogResponse.ResponseBody)).To(ContainSubstring(`"maintenance_info":{}`))
			})
		})

	})

})
