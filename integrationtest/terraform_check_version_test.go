package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Terraform check version", func() {
	const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
	const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"

	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-check-version")

		session = testHelper.StartBroker()
	})

	AfterEach(func() {
		session.Terminate().Wait()
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

	lastOperationMessage := func(serviceInstanceGUID string) string {
		var tfDeploymentReceiver models.TerraformDeployment
		Expect(testHelper.DBConn().Where("id = ?", fmt.Sprintf("%s", serviceInstanceGUID)).First(&tfDeploymentReceiver).Error).NotTo(HaveOccurred())
		return tfDeploymentReceiver.LastOperationMessage
	}

	Describe("Update", func() {
		When("Default Terraform version greater than instance", func() {
			It("returns an error", func() {
				By("provisioning a service instance at 0.12")
				serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

				By("updating the brokerpak and restarting the broker")
				session.Terminate().Wait()
				testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-check-version-updated")
				session = testHelper.StartBroker()

				By("running 'cf update-service'")
				updateResponse := testHelper.Client().Update(serviceInstance.GUID, serviceOfferingGUID, servicePlanGUID, requestID(), nil)
				Expect(updateResponse.Error).NotTo(HaveOccurred())
				Expect(updateResponse.StatusCode).To(Equal(http.StatusAccepted))
				Expect(testHelper.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Failed))

				By("observing that the TF version remains the same in the state file")
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

				By("observing that the destroy failed due to mismatched TF versions")
				Expect(testHelper.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Failed))
				Eventually(lastOperationMessage("tf:" + serviceInstance.GUID + ":")).WithTimeout(10 * time.Second).Should(Equal("apply attempted with a newer version of terraform than the state"))
			})
		})
	})

	Describe("Unbind", func() {
		When("Default Terraform version greater than instance", func() {
			It("returns an error", func() {
				By("provisioning a service instance at 0.12")
				serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

				By("creating a binding")
				bindGUID, _ := testHelper.CreateBinding(serviceInstance)

				By("updating the brokerpak and restarting the broker")
				session.Terminate().Wait()
				testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-check-version-updated")

				session = testHelper.StartBroker()

				By("deleting the instance binding")
				deleteBindResponse := testHelper.Client().Unbind(serviceInstance.GUID, bindGUID, serviceOfferingGUID, servicePlanGUID, requestID())
				Expect(deleteBindResponse.StatusCode).To(Equal(http.StatusInternalServerError))

				By("observing that the destroy failed due to mismatched TF versions")
				Eventually(session).WithTimeout(10 * time.Second).Should(gbytes.Say("apply attempted with a newer version of terraform than the state"))
				Eventually(lastOperationMessage("tf:" + serviceInstance.GUID + ":" + bindGUID)).WithTimeout(10 * time.Second).Should(Equal("apply attempted with a newer version of terraform than the state"))
			})
		})
	})

	Describe("Delete", func() {
		When("Default Terraform version greater than instance", func() {
			It("returns an error", func() {
				By("provisioning a service instance at 0.12")
				serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

				By("updating the brokerpak and restarting the broker")
				session.Terminate().Wait()
				testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-check-version-updated")

				session = testHelper.StartBroker()

				By("deleting the service instance")
				deleteBindResponse := testHelper.Client().Deprovision(serviceInstance.GUID, serviceOfferingGUID, servicePlanGUID, requestID())
				Expect(deleteBindResponse.Error).NotTo(HaveOccurred())
				Expect(deleteBindResponse.StatusCode).To(Equal(http.StatusAccepted))
				Expect(testHelper.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Failed))
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

				By("observing that the destroy failed due to mismatched TF versions")
				Expect(testHelper.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Failed))
				Eventually(lastOperationMessage("tf:" + serviceInstance.GUID + ":")).WithTimeout(10 * time.Second).Should(Equal("apply attempted with a newer version of terraform than the state"))
			})
		})
	})

})
