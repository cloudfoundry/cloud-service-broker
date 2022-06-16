package integrationtest_test

import (
	"fmt"
	"net/http"

	"github.com/pborman/uuid"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Terraform Module Upgrade", func() {
	const (
		serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		servicePlanGUID     = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
	)

	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	instanceTerraformStateVersion := func(serviceInstanceGUID string) string {
		return terraformStateVersion(fmt.Sprintf("tf:%s:", serviceInstanceGUID), testHelper)
	}

	bindingTerraformStateVersion := func(serviceInstanceGUID, bindingGUID string) string {
		return terraformStateVersion(fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, bindingGUID), testHelper)
	}

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-module-upgrade")

		session = testHelper.StartBroker()

		DeferCleanup(func() {
			session.Terminate().Wait()
		})
	})

	Context("TF Upgrades are enabled", func() {
		It("runs 'terraform apply' at each version in the upgrade path", func() {
			By("provisioning a service instance at 0.12")
			serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

			By("creating service bindings")
			firstBindGUID := uuid.New()
			firstBindResponse := testHelper.Client().Bind(serviceInstance.GUID, firstBindGUID, serviceOfferingGUID, servicePlanGUID, requestID(), nil)
			Expect(firstBindResponse.Error).NotTo(HaveOccurred())
			Expect(firstBindResponse.StatusCode).To(Equal(http.StatusCreated))

			secondBindGUID := uuid.New()
			secondBindResponse := testHelper.Client().Bind(serviceInstance.GUID, secondBindGUID, serviceOfferingGUID, servicePlanGUID, requestID(), nil)
			Expect(secondBindResponse.Error).NotTo(HaveOccurred())
			Expect(secondBindResponse.StatusCode).To(Equal(http.StatusCreated))

			By("updating the brokerpak and restarting the broker")
			session.Terminate().Wait()
			testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-module-upgrade-updated")

			session = testHelper.StartBroker(
				"TERRAFORM_UPGRADES_ENABLED=true",
				"BROKERPAK_UPDATES_ENABLED=true",
			)

			By("validating old state")
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, firstBindGUID)).To(Equal("0.12.21"))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, secondBindGUID)).To(Equal("0.12.21"))

			By("running 'cf update-service'")
			testHelper.UpgradeService(serviceInstance, domain.PreviousValues{PlanID: servicePlanGUID}, domain.MaintenanceInfo{Version: "1.1.6"})

			By("observing that the instance TF state file has been updated to the latest version")
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal("1.1.6"))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, firstBindGUID)).To(Equal("1.1.6"))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, secondBindGUID)).To(Equal("1.1.6"))
		})
	})
})
