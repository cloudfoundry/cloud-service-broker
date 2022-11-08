package integrationtest_test

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Terraform block action before upgrade", func() {
	const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
	const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"

	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	BeforeEach(func() {
		brokerpak = must(testdrive.BuildBrokerpak(csb, fixtures("terraform-block-action-before-upgrade")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	terraformStateVersion := func(serviceInstanceGUID string) string {
		var tfDeploymentReceiver models.TerraformDeployment
		Expect(dbConn.Where("id = ?", fmt.Sprintf("tf:%s:", serviceInstanceGUID)).First(&tfDeploymentReceiver).Error).NotTo(HaveOccurred())
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
				serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

				By("updating the brokerpak and restarting the broker")
				Expect(broker.Stop()).To(Succeed())
				_, err := testdrive.BuildBrokerpak(csb, fixtures("terraform-block-action-before-upgrade-updated"), testdrive.WithDirectory(brokerpak))
				Expect(err).NotTo(HaveOccurred())

				broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

				By("creating a binding")
				_, err = broker.CreateBinding(serviceInstance)
				Expect(err).To(MatchError(ContainSubstring("operation attempted with newer version of Terraform than current state, upgrade the service before retrying operation")))
				Expect(broker.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Succeeded))
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))
			})
		})
	})

	Describe("Unbind", func() {
		When("Default Terraform version greater than instance", func() {
			It("returns an error", func() {
				By("provisioning a service instance at 0.12")
				serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

				By("creating a binding")
				binding := must(broker.CreateBinding(serviceInstance))

				By("updating the brokerpak and restarting the broker")
				Expect(broker.Stop()).To(Succeed())
				_, err := testdrive.BuildBrokerpak(csb, fixtures("terraform-block-action-before-upgrade-updated"), testdrive.WithDirectory(brokerpak))
				Expect(err).NotTo(HaveOccurred())

				broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

				By("deleting the instance binding")
				err = broker.DeleteBinding(serviceInstance, binding.GUID)
				Expect(err).To(MatchError(ContainSubstring("operation attempted with newer version of Terraform than current state, upgrade the service before retrying operation")))
				Expect(broker.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Succeeded))
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))
			})
		})
	})

	Describe("Delete", func() {
		When("Default Terraform version greater than instance", func() {
			It("returns an error", func() {
				By("provisioning a service instance at 0.12")
				serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))

				By("updating the brokerpak and restarting the broker")
				Expect(broker.Stop()).To(Succeed())
				_, err := testdrive.BuildBrokerpak(csb, fixtures("terraform-block-action-before-upgrade-updated"), testdrive.WithDirectory(brokerpak))
				Expect(err).NotTo(HaveOccurred())

				broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

				By("deleting the service instance")
				err = broker.Deprovision(serviceInstance)
				Expect(err).To(MatchError(ContainSubstring("operation attempted with newer version of Terraform than current state, upgrade the service before retrying operation")))
				Expect(broker.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Succeeded))
				Expect(terraformStateVersion(serviceInstance.GUID)).To(Equal("0.12.21"))
			})
		})
	})
})
