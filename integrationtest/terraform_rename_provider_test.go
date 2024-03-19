package integrationtest_test

import (
	"github.com/cloudfoundry/cloud-service-broker/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v10/domain"
)

var _ = Describe("Terraform Rename Provider", func() {
	var (
		brokerpak string
		broker    *testdrive.Broker
	)
	const terraformVersion = "1.6.2"

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("terraform-rename-provider")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	It("can provision when provider is renamed", func() {
		const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
		serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))

		Expect(broker.Stop()).To(Succeed())
		must(packer.BuildBrokerpak(csb, fixtures("terraform-rename-provider"), packer.WithDirectory(brokerpak)))

		broker = must(testdrive.StartBroker(
			csb, brokerpak, database,
			testdrive.WithEnv("TERRAFORM_UPGRADES_ENABLED=true", "BROKERPAK_UPDATES_ENABLED=true"),
			testdrive.WithOutputs(GinkgoWriter, GinkgoWriter),
		))

		By("running 'cf update-service'")
		Expect(broker.UpgradeService(
			serviceInstance, terraformVersion,
			testdrive.WithUpgradePreviousValues(domain.PreviousValues{
				PlanID:          servicePlanGUID,
				MaintenanceInfo: &domain.MaintenanceInfo{Version: terraformVersion},
			}),
			testdrive.WithUpgradeParams(`{"alpha_input":"quz"}`),
		)).To(Succeed())
	})

	It("can delete instance when provider is renamed", func() {
		const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
		serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))

		Expect(broker.Stop()).To(Succeed())
		must(packer.BuildBrokerpak(csb, fixtures("terraform-rename-provider"), packer.WithDirectory(brokerpak)))

		broker = must(testdrive.StartBroker(
			csb, brokerpak, database,
			testdrive.WithEnv("TERRAFORM_UPGRADES_ENABLED=true", "BROKERPAK_UPDATES_ENABLED=true"),
			testdrive.WithOutputs(GinkgoWriter, GinkgoWriter),
		))

		By("running 'cf delete-service'")
		Expect(broker.Deprovision(serviceInstance)).To(Succeed())
	})

	It("can delete binding when provider is renamed", func() {
		const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
		serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
		binding := must(broker.CreateBinding(serviceInstance))

		Expect(broker.Stop()).To(Succeed())
		must(packer.BuildBrokerpak(csb, fixtures("terraform-rename-provider"), packer.WithDirectory(brokerpak)))

		broker = must(testdrive.StartBroker(
			csb, brokerpak, database,
			testdrive.WithEnv("TERRAFORM_UPGRADES_ENABLED=true", "BROKERPAK_UPDATES_ENABLED=true"),
			testdrive.WithOutputs(GinkgoWriter, GinkgoWriter),
		))

		By("running 'cf delete-binding'")
		Expect(broker.DeleteBinding(serviceInstance, binding.GUID)).To(Succeed())
	})
})
