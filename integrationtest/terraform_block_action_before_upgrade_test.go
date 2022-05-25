package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pborman/uuid"

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
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-block-action-before-upgrade")

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

	Describe("Bind", func() {
		When("Default Terraform version greater than instance", func() {
			It("returns an error", func() {
				By("provisioning a service instance at 0.12")
				serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

				By("updating the brokerpak and restarting the broker")
				session.Terminate().Wait()
				testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-block-action-before-upgrade-updated")

				session = testHelper.StartBroker()

				By("creating a binding")
				bindResponse := testHelper.Client().Bind(serviceInstance.GUID, uuid.New(), serviceInstance.ServiceOfferingGUID, serviceInstance.ServicePlanGUID, uuid.New(), []byte(""))
				Expect(bindResponse.Error).NotTo(HaveOccurred())
				Expect(bindResponse.StatusCode).To(Equal(http.StatusInternalServerError))
				Expect(bindResponse.ResponseBody).To(ContainSubstring("apply attempted with a newer version of terraform than the state"))
				Expect(testHelper.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Succeeded))
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

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
				testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-block-action-before-upgrade-updated")

				session = testHelper.StartBroker()

				By("deleting the instance binding")
				unBindResponse := testHelper.Client().Unbind(serviceInstance.GUID, bindGUID, serviceOfferingGUID, servicePlanGUID, requestID())
				Expect(unBindResponse.Error).NotTo(HaveOccurred())
				Expect(unBindResponse.StatusCode).To(Equal(http.StatusInternalServerError))
				Expect(unBindResponse.ResponseBody).To(ContainSubstring("apply attempted with a newer version of terraform than the state"))
				Expect(testHelper.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Succeeded))
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))
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
				testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "terraform-block-action-before-upgrade-updated")

				session = testHelper.StartBroker()

				By("deleting the service instance")
				deleteBindResponse := testHelper.Client().Deprovision(serviceInstance.GUID, serviceOfferingGUID, servicePlanGUID, requestID())
				Expect(deleteBindResponse.Error).NotTo(HaveOccurred())
				Expect(deleteBindResponse.StatusCode).To(Equal(http.StatusInternalServerError))
				Expect(deleteBindResponse.ResponseBody).To(ContainSubstring("apply attempted with a newer version of terraform than the state"))
				Expect(testHelper.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Succeeded))
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))
			})
		})
	})

})
