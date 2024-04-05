package integrationtest_test

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v11/domain"
)

var _ = Describe("Terraform Upgrade", func() {
	const (
		serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		servicePlanGUID     = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
		startingVersion     = "1.6.0"
		endingVersion       = "1.6.2"
	)

	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	terraformStateOutputValue := func(deploymentID, outputName string) any {
		var tfDeploymentReceiver models.TerraformDeployment
		Expect(dbConn.Where("id = ?", deploymentID).First(&tfDeploymentReceiver).Error).NotTo(HaveOccurred())
		var workspaceReceiver struct {
			State []byte `json:"tfstate"`
		}
		Expect(json.Unmarshal(tfDeploymentReceiver.Workspace, &workspaceReceiver)).NotTo(HaveOccurred())
		var stateReceiver struct {
			Outputs map[string]struct {
				Type  string `json:"type"`
				Value any    `json:"value"`
			} `json:"outputs"`
		}
		Expect(json.Unmarshal(workspaceReceiver.State, &stateReceiver)).NotTo(HaveOccurred())
		Expect(stateReceiver.Outputs).To(HaveKey(outputName), "could not find output with this name")
		return stateReceiver.Outputs[outputName].Value
	}

	instanceTerraformStateVersion := func(serviceInstanceGUID string) string {
		return terraformStateVersion(fmt.Sprintf("tf:%s:", serviceInstanceGUID))
	}

	bindingTerraformStateVersion := func(serviceInstanceGUID, bindingGUID string) string {
		return terraformStateVersion(fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, bindingGUID))
	}

	instanceTerraformStateOutputValue := func(serviceInstanceGUID string) int {
		val := terraformStateOutputValue(fmt.Sprintf("tf:%s:", serviceInstanceGUID), "provision_output")
		return int(val.(float64))
	}

	bindingTerraformStateOutputValue := func(serviceInstanceGUID, bindingGUID string) int {
		val := terraformStateOutputValue(fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, bindingGUID), "provision_output")
		return int(val.(float64))
	}

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("terraform-upgrade")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	Context("TF Upgrades are enabled", func() {
		It("runs 'terraform apply' at each version in the upgrade path", func() {
			By("provisioning a service instance at 1.6.0")
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal(startingVersion))

			By("creating service bindings")
			firstBinding := must(broker.CreateBinding(serviceInstance))
			secondBinding := must(broker.CreateBinding(serviceInstance))

			By("updating the brokerpak and restarting the broker")
			Expect(broker.Stop()).To(Succeed())
			must(packer.BuildBrokerpak(csb, fixtures("terraform-upgrade-updated"), packer.WithDirectory(brokerpak)))

			broker = must(testdrive.StartBroker(
				csb, brokerpak, database,
				testdrive.WithEnv("TERRAFORM_UPGRADES_ENABLED=true"),
				testdrive.WithOutputs(GinkgoWriter, GinkgoWriter),
			))

			By("validating old state")
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal(startingVersion))
			Expect(instanceTerraformStateOutputValue(serviceInstance.GUID)).To(BeElementOf(1, 2))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, firstBinding.GUID)).To(Equal(startingVersion))
			Expect(bindingTerraformStateOutputValue(serviceInstance.GUID, firstBinding.GUID)).To(BeElementOf(1, 2))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, secondBinding.GUID)).To(Equal(startingVersion))
			Expect(bindingTerraformStateOutputValue(serviceInstance.GUID, secondBinding.GUID)).To(BeElementOf(1, 2))

			By("running 'cf upgrade-service'")
			Expect(broker.UpgradeService(serviceInstance, endingVersion, testdrive.WithUpgradePreviousValues(domain.PreviousValues{PlanID: servicePlanGUID}))).To(Succeed())

			By("observing that the instance TF state file has been updated to the latest version")
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal(endingVersion))
			By("observing that the instance TF state file has updated output value")
			Expect(instanceTerraformStateOutputValue(serviceInstance.GUID)).To(BeElementOf(3, 4))

			By("observing that the binding TF state file has been updated to the latest version")
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, firstBinding.GUID)).To(Equal(endingVersion))
			Expect(bindingTerraformStateVersion(serviceInstance.GUID, secondBinding.GUID)).To(Equal(endingVersion))
			By("observing that the binding TF state file has updated output value")
			Expect(bindingTerraformStateOutputValue(serviceInstance.GUID, firstBinding.GUID)).To(BeElementOf(3, 4))
			Expect(bindingTerraformStateOutputValue(serviceInstance.GUID, secondBinding.GUID)).To(BeElementOf(3, 4))

			By("updating the service after the upgrade")
			Expect(broker.UpdateService(
				serviceInstance,
				testdrive.WithUpdateParams(`{"alpha_input":"foo"}`),
				testdrive.WithUpdatePreviousValues(domain.PreviousValues{MaintenanceInfo: &domain.MaintenanceInfo{Version: endingVersion}}),
			)).To(Succeed())
		})
	})

	Context("TF Upgrades are disabled", func() {
		It("does not upgrade the instance", func() {
			By("provisioning a service instance at 1.6.0")
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal(startingVersion))

			By("updating the brokerpak and restarting the broker")
			Expect(broker.Stop()).To(Succeed())
			must(packer.BuildBrokerpak(csb, fixtures("terraform-upgrade-updated"), packer.WithDirectory(brokerpak)))

			broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

			By("seeing 'cf update-service' fail")
			Expect(broker.UpdateService(serviceInstance)).To(MatchError(ContainSubstring("terraform version check failed: operation attempted with newer version of Terraform than current state, upgrade the service before retrying operation")))

			By("observing that the TF version remains the same in the state file")
			Expect(instanceTerraformStateVersion(serviceInstance.GUID)).To(Equal(startingVersion))
		})
	})
})

func terraformStateVersion(deploymentID string) string {
	var tfDeploymentReceiver models.TerraformDeployment
	Expect(dbConn.Where("id = ?", deploymentID).First(&tfDeploymentReceiver).Error).NotTo(HaveOccurred())
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
