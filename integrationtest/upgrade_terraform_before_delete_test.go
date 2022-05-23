package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("upgrade terraform before delete", func() {
	const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
	const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"

	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "upgrade-terraform-before-delete")

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

	Context("Terraform Upgrades are enabled", func() {
		Describe("service instance created with Terraform 0.12", func() {
			It("upgrades the terraform to latest before deleting", func() {
				By("provisioning a service instance at 0.12")
				serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

				By("updating the brokerpak and restarting the broker")
				session.Terminate().Wait()
				testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "upgrade-terraform-before-delete-updated")

				session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=true")

				By("deleting the service instance")
				testHelper.Deprovision(serviceInstance)

				By("observing that the TF version has been updated to latest before destroy")
				Expect(session).To(gbytes.Say("versions/0.13.7/terraform\",\"apply\""))
				Expect(session).To(gbytes.Say("versions/0.14.9/terraform\",\"apply\""))
				Expect(session).To(gbytes.Say("versions/1.0.10/terraform\",\"apply\""))
				Expect(session).To(gbytes.Say("versions/1.1.6/terraform\",\"apply\""))
			})
		})
	})

	Context("Terraform upgrades are disabled", func() {
		Describe("existing service created with Terraform 0.12", func() {
			It("fails to delete service", func() {
				By("provisioning a service instance at 0.12")
				serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

				By("updating the brokerpak and restarting the broker")
				session.Terminate().Wait()
				testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "upgrade-terraform-before-delete-updated")

				session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=false")

				By("deleting the service instance")
				deleteBindResponse := testHelper.Client().Deprovision(serviceInstance.GUID, serviceOfferingGUID, servicePlanGUID, requestID())
				Expect(deleteBindResponse.Error).NotTo(HaveOccurred())
				Expect(deleteBindResponse.StatusCode).To(Equal(http.StatusAccepted))
				Expect(testHelper.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Failed))
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))
			})
		})
	})
})
