package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/onsi/gomega/gbytes"

	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"

	"github.com/cloudfoundry/cloud-service-broker/db_service/models"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unbind", func() {
	const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
	const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"

	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "brokerpak-terraform-0.12")

		session = testHelper.StartBroker()
	})

	AfterEach(func() {
		session.Terminate()
		testHelper.Restore()
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
		Describe("existing binding created with Terraform 0.12", func() {
			It("upgrades the terraform to latest", func() {
				By("provisioning a service instance at 0.12")
				serviceInstanceGUID := uuid.New()
				provisionResponse := testHelper.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), nil)
				Expect(provisionResponse.Error).NotTo(HaveOccurred())
				Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))

				Eventually(pollLastOperation(testHelper, serviceInstanceGUID), time.Minute*2, lastOperationPollingFrequency).ShouldNot(Equal(domain.InProgress))
				Expect(pollLastOperation(testHelper, serviceInstanceGUID)()).To(Equal(domain.Succeeded))
				Expect(terraformStateVersion(serviceInstanceGUID)).To(Equal("0.12.21"))

				By("creating a binding")
				bindGUID := uuid.New()
				bindResponse := testHelper.Client().Bind(serviceInstanceGUID, bindGUID, serviceOfferingGUID, servicePlanGUID, requestID(), nil)

				Expect(bindResponse.Error).NotTo(HaveOccurred())
				Expect(bindResponse.StatusCode).To(Equal(http.StatusCreated))

				By("updating the brokerpak and restarting the broker")
				session.Terminate()
				testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "brokerpak-terraform-upgrade")

				session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=true")

				By("deleting the instance binding")
				deleteBindResponse := testHelper.Client().Unbind(serviceInstanceGUID, bindGUID, serviceOfferingGUID, servicePlanGUID, requestID())

				Expect(deleteBindResponse.Error).NotTo(HaveOccurred())
				Expect(deleteBindResponse.StatusCode).To(Equal(http.StatusOK))

				By("observing that the TF state file has been updated to the latest version before destroy")
				Expect(session).To(gbytes.Say("versions/0.13.7/terraform\",\"apply\""))
				Expect(session).To(gbytes.Say("versions/0.14.9/terraform\",\"apply\""))
				Expect(session).To(gbytes.Say("versions/1.0.10/terraform\",\"apply\""))
				Expect(session).To(gbytes.Say("versions/1.1.6/terraform\",\"apply\""))

			})
		})
	})

	Context("Terraform upgrades are disabled", func() {
		Describe("existing binding created with Terraform 0.12", func() {
			FIt("fails to delete binding", func() {
				By("provisioning a service instance at 0.12")
				serviceInstanceGUID := uuid.New()
				provisionResponse := testHelper.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), nil)
				Expect(provisionResponse.Error).NotTo(HaveOccurred())
				Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))

				Eventually(pollLastOperation(testHelper, serviceInstanceGUID), time.Minute*2, lastOperationPollingFrequency).ShouldNot(Equal(domain.InProgress))
				Expect(pollLastOperation(testHelper, serviceInstanceGUID)()).To(Equal(domain.Succeeded))
				Expect(terraformStateVersion(serviceInstanceGUID)).To(Equal("0.12.21"))

				By("creating a binding")
				bindGUID := uuid.New()
				bindResponse := testHelper.Client().Bind(serviceInstanceGUID, bindGUID, serviceOfferingGUID, servicePlanGUID, requestID(), nil)

				Expect(bindResponse.Error).NotTo(HaveOccurred())
				Expect(bindResponse.StatusCode).To(Equal(http.StatusCreated))

				By("updating the brokerpak and restarting the broker")
				session.Terminate()
				testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "brokerpak-terraform-upgrade")

				session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=false")

				By("deleting the instance binding")
				deleteBindResponse := testHelper.Client().Unbind(serviceInstanceGUID, bindGUID, serviceOfferingGUID, servicePlanGUID, requestID())

				Expect(deleteBindResponse.StatusCode).To(Equal(http.StatusInternalServerError))

				By("observing that the destroy failed due to mismatched TF versions")
				Expect(session).To(gbytes.Say("apply attempted with a newer version of terraform than the state"))

			})
		})
	})

})
