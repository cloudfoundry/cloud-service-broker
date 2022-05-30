package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Terraform 0.12 Upgrade", func() {
	const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
	const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"

	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-upgrade-0.12")

		session = testHelper.StartBroker()
	})

	AfterEach(func() {
		session.Terminate()
	})

	terraformStateVersion := func(serviceInstanceGUID string) string {
		var tfDeploymentReceiver models.TerraformDeployment
		Expect(testHelper.DBConn().Where("id = ?", fmt.Sprintf("tf:%s:", serviceInstanceGUID)).First(&tfDeploymentReceiver).Error).NotTo(HaveOccurred())
		var workspaceReceiver struct {
			State []byte `json:"tfstate"`
		}
		Expect(json.Unmarshal(tfDeploymentReceiver.Workspace, &workspaceReceiver)).NotTo(HaveOccurred())
		var stateReceiver struct {
			Version string `json:"terraform_version"`
		}
		Expect(json.Unmarshal(workspaceReceiver.State, &stateReceiver)).NotTo(HaveOccurred())
		return stateReceiver.Version
	}

	Context("TF Upgrades are enabled", func() {
		It("runs 'terraform apply' at each version in the upgrade path", func() {
			By("provisioning a service instance at 0.12")
			serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
			Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

			By("updating the brokerpak and restarting the broker")
			session.Terminate().Wait()
			testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-upgrade-0.12-updated")

			session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=true")

			By("running 'cf update-service'")
			newMI := domain.MaintenanceInfo{
				Version: "1.1.6",
			}
			testHelper.UpgradeService(serviceInstance, domain.PreviousValues{PlanID: servicePlanGUID}, newMI)

			By("observing that the TF state file has been updated to the latest version")
			Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("1.1.6"))
		})
	})

	Context("TF Upgrades are disabled", func() {
		It("does not upgrade the instance", func() {
			By("provisioning a service instance at 0.12")
			serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
			Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

			By("updating the brokerpak and restarting the broker")
			session.Terminate().Wait()
			testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-upgrade")
			session = testHelper.StartBroker()

			By("running 'cf update-service'")
			updateResponse := testHelper.Client().Update(serviceInstance.GUID, serviceOfferingGUID, servicePlanGUID, requestID(), nil, domain.PreviousValues{}, nil)
			Expect(updateResponse.Error).NotTo(HaveOccurred())
			Expect(updateResponse.StatusCode).To(Equal(http.StatusAccepted))
			Expect(testHelper.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Failed))

			By("observing that the TF version remains the same in the state file")
			Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))
		})
	})
})
