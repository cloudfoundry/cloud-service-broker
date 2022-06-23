package integrationtest_test

import (
	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Terraform Rename Provider", func() {
	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-rename-provider")
		session = testHelper.StartBroker()
	})

	AfterEach(func() {
		session.Terminate().Wait()
	})

	It("can provision when provider is renamed", func() {
		const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
		serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)

		session.Terminate().Wait()
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-rename-provider")
		session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=true", "BROKERPAK_UPDATES_ENABLED=true")

		By("running 'cf update-service'")
		testHelper.UpgradeService(
			serviceInstance,
			domain.PreviousValues{
				PlanID:          servicePlanGUID,
				MaintenanceInfo: &domain.MaintenanceInfo{Version: "0.13.7"},
			},
			domain.MaintenanceInfo{Version: "0.13.7"}, `{"alpha_input":"quz"}`,
		)
	})

	It("can delete instance when provider is renamed", func() {
		const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
		serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
		session.Terminate().Wait()

		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-rename-provider")
		session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=true", "BROKERPAK_UPDATES_ENABLED=true")

		By("running 'cf delete-service'")
		testHelper.Deprovision(serviceInstance)
	})

	It("can delete binding when provider is renamed", func() {
		const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
		serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)

		bindingGUID, _ := testHelper.CreateBinding(serviceInstance)

		session.Terminate().Wait()

		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-rename-provider")
		session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=true", "BROKERPAK_UPDATES_ENABLED=true")

		By("running 'cf delete-binding'")
		testHelper.DeleteBinding(serviceInstance, bindingGUID)
	})
})
