package integrationtest_test

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Terraform Module Upgrade", func() {
	const (
		serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		servicePlanGUID     = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
	)

	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	instanceTerraformStateVersion := func(serviceInstanceGUID string) string {
		return terraformStateVersion(fmt.Sprintf("tf:%s:", serviceInstanceGUID))
	}

	bindingTerraformStateVersion := func(serviceInstanceGUID, bindingGUID string) string {
		return terraformStateVersion(fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, bindingGUID))
	}

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("terraform-module-upgrade")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	Context("TF Upgrades are enabled", func() {
		It("runs 'terraform apply' at each version in the upgrade path", func() {
			By("provisioning a service instance at 0.12")
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

			By("creating service bindings")
			firstBinding := must(broker.CreateBinding(serviceInstance))
			secondBinding := must(broker.CreateBinding(serviceInstance))

			By("updating the brokerpak and restarting the broker")
			Expect(broker.Stop()).To(Succeed())
			must(packer.BuildBrokerpak(csb, fixtures("terraform-module-upgrade-updated"), packer.WithDirectory(brokerpak)))

			broker = must(testdrive.StartBroker(
				csb, brokerpak, database,
				testdrive.WithOutputs(GinkgoWriter, GinkgoWriter),
				testdrive.WithEnv("TERRAFORM_UPGRADES_ENABLED=true", "BROKERPAK_UPDATES_ENABLED=true"),
			))

			By("validating old state")
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, firstBinding.GUID)).To(Equal("0.12.21"))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, secondBinding.GUID)).To(Equal("0.12.21"))

			By("running 'cf update-service'")
			Expect(broker.UpgradeService(serviceInstance, "1.1.6", testdrive.WithUpgradePreviousValues(domain.PreviousValues{PlanID: servicePlanGUID}))).To(Succeed())

			By("observing that the instance TF state file has been updated to the latest version")
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal("1.1.6"))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, firstBinding.GUID)).To(Equal("1.1.6"))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, secondBinding.GUID)).To(Equal("1.1.6"))
		})
	})
})
