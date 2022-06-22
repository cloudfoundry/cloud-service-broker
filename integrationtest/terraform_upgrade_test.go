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

var _ = Describe("Terraform Upgrade", func() {
	const (
		serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		servicePlanGUID     = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
		startingVersion     = "0.12.21"
		endingVersion       = "1.2.3"
	)

	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	terraformStateOutputValue := func(deploymentID string, testHelper *helper.TestHelper) int {
		var tfDeploymentReceiver models.TerraformDeployment
		Expect(testHelper.DBConn().Where("id = ?", deploymentID).First(&tfDeploymentReceiver).Error).NotTo(HaveOccurred())
		var workspaceReceiver struct {
			State []byte `json:"tfstate"`
		}
		Expect(json.Unmarshal(tfDeploymentReceiver.Workspace, &workspaceReceiver)).NotTo(HaveOccurred())
		var stateReceiver struct {
			Outputs map[string]struct {
				Type  string      `json:"type"`
				Value interface{} `json:"value"`
			} `json:"outputs"`
		}
		Expect(json.Unmarshal(workspaceReceiver.State, &stateReceiver)).NotTo(HaveOccurred())

		var result float64
		for _, value := range stateReceiver.Outputs {
			result = value.Value.(float64)
		}
		return int(result)
	}

	instanceTerraformStateVersion := func(serviceInstanceGUID string) string {
		return terraformStateVersion(fmt.Sprintf("tf:%s:", serviceInstanceGUID), testHelper)
	}

	bindingTerraformStateVersion := func(serviceInstanceGUID, bindingGUID string) string {
		return terraformStateVersion(fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, bindingGUID), testHelper)
	}

	instanceTerraformStateOutputValue := func(serviceInstanceGUID string) int {
		return terraformStateOutputValue(fmt.Sprintf("tf:%s:", serviceInstanceGUID), testHelper)
	}

	bindingTerraformStateOutputValue := func(serviceInstanceGUID, bindingGUID string) int {
		return terraformStateOutputValue(fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, bindingGUID), testHelper)
	}

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-upgrade")
		session = testHelper.StartBroker()

		DeferCleanup(func() {
			session.Terminate().Wait()
		})
	})

	Context("TF Upgrades are enabled", func() {
		It("runs 'terraform apply' at each version in the upgrade path", func() {
			By("provisioning a service instance at 0.13")
			serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal(startingVersion))

			By("creating service bindings")
			firstBindGUID, _ := testHelper.CreateBinding(serviceInstance)
			secondBindGUID, _ := testHelper.CreateBinding(serviceInstance)

			By("updating the brokerpak and restarting the broker")
			session.Terminate().Wait()
			testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-upgrade-updated")

			session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=true")

			By("validating old state")
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal(startingVersion))
			Expect(instanceTerraformStateOutputValue(serviceInstance.GUID)).To(BeElementOf(1, 2))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, firstBindGUID)).To(Equal(startingVersion))
			Expect(bindingTerraformStateOutputValue(serviceInstance.GUID, firstBindGUID)).To(BeElementOf(1, 2))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, secondBindGUID)).To(Equal(startingVersion))
			Expect(bindingTerraformStateOutputValue(serviceInstance.GUID, secondBindGUID)).To(BeElementOf(1, 2))

			By("running 'cf update-service'")
			testHelper.UpgradeService(serviceInstance, domain.PreviousValues{PlanID: servicePlanGUID}, domain.MaintenanceInfo{Version: endingVersion})

			By("observing that the instance TF state file has been updated to the latest version")
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal(endingVersion))
			By("observing that the instance TF state file has updated output value")
			Expect(instanceTerraformStateOutputValue(serviceInstance.GUID)).To(BeElementOf(3, 4))

			By("observing that the binding TF state file has been updated to the latest version")
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, firstBindGUID)).To(Equal(endingVersion))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, secondBindGUID)).To(Equal(endingVersion))
			By("observing that the binding TF state file has updated output value")
			Expect(bindingTerraformStateOutputValue(serviceInstance.GUID, firstBindGUID)).To(BeElementOf(3, 4))
			Expect(bindingTerraformStateOutputValue(serviceInstance.GUID, secondBindGUID)).To(BeElementOf(3, 4))

			By("updating the service after the upgrade")
			testHelper.UpdateService(serviceInstance, `{"alpha_input":"foo"}`)
		})
	})

	Context("TF Upgrades are disabled", func() {
		It("does not upgrade the instance", func() {
			By("provisioning a service instance at 0.13")
			serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal(startingVersion))

			By("updating the brokerpak and restarting the broker")
			session.Terminate().Wait()
			testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-upgrade-updated")
			session = testHelper.StartBroker()

			By("seeing 'cf update-service' fail")
			updateResponse := testHelper.Client().Update(serviceInstance.GUID, serviceOfferingGUID, servicePlanGUID, requestID(), nil, domain.PreviousValues{}, nil)
			Expect(updateResponse.Error).NotTo(HaveOccurred())
			Expect(updateResponse.StatusCode).To(Equal(http.StatusInternalServerError))
			Expect(updateResponse.ResponseBody).To(ContainSubstring("terraform version check failed: operation attempted with newer version of Terraform than current state, upgrade the service before retrying operation"))

			By("observing that the TF version remains the same in the state file")
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal(startingVersion))
		})
	})
})

func terraformStateVersion(deploymentID string, testHelper *helper.TestHelper) string {
	var tfDeploymentReceiver models.TerraformDeployment
	Expect(testHelper.DBConn().Where("id = ?", deploymentID).First(&tfDeploymentReceiver).Error).NotTo(HaveOccurred())
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
