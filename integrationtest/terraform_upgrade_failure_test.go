package integrationtest_test

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/pivotal-cf/brokerapi/v12/domain"

	"github.com/cloudfoundry/cloud-service-broker/v2/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TF Upgrade Failure", func() {
	const (
		serviceOfferingGUID = "10557d15-dd47-40e6-ab4f-53fbe81e3022"
		servicePlanGUID     = "e328fdae-d97c-43b1-a1a7-0f8e961d1d0c"
		startingVersion     = "1.6.0"
		endingVersion       = "1.6.2"
	)

	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	tfStateOutputValue := func(deploymentID, outputName string) any {
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

	instanceTFStateVersion := func(serviceInstanceGUID string) string {
		return terraformStateVersion(fmt.Sprintf("tf:%s:", serviceInstanceGUID))
	}

	bindingTFStateVersion := func(serviceInstanceGUID, bindingGUID string) string {
		return terraformStateVersion(fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, bindingGUID))
	}

	instanceTFStateOutputValue := func(serviceInstanceGUID string) int {
		val := tfStateOutputValue(fmt.Sprintf("tf:%s:", serviceInstanceGUID), "provision_output")
		return int(val.(float64))
	}

	bindingTFStateOutputValue := func(serviceInstanceGUID, bindingGUID string) int {
		val := tfStateOutputValue(fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, bindingGUID), "provision_output")
		return int(val.(float64))
	}

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("terraform-upgrade-failure")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	It("can recover an upgrade after a failure", func() {
		By("provisioning a service instance at 1.6.0")
		serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(map[string]any{"max": 2})))
		Expect(instanceTFStateVersion(serviceInstance.GUID)).To(Equal(startingVersion))

		By("creating a service binding")
		binding := must(broker.CreateBinding(serviceInstance))

		By("updating the brokerpak and restarting the broker")
		Expect(broker.Stop()).To(Succeed())
		must(packer.BuildBrokerpak(csb, fixtures("terraform-upgrade-failure-updated"), packer.WithDirectory(brokerpak)))

		broker = must(testdrive.StartBroker(
			csb, brokerpak, database,
			testdrive.WithEnv("TERRAFORM_UPGRADES_ENABLED=true"),
			testdrive.WithOutputs(GinkgoWriter, GinkgoWriter),
		))

		By("validating the service instance and binding are at the starting TF version")
		Expect(instanceTFStateVersion(serviceInstance.GUID)).To(Equal(startingVersion))
		Expect(bindingTFStateVersion(serviceInstance.GUID, binding.GUID)).To(Equal(startingVersion))

		By("observing the instance and binding outputs are in the correct range")
		Expect(instanceTFStateOutputValue(serviceInstance.GUID)).To(BeElementOf(1, 2))
		Expect(bindingTFStateOutputValue(serviceInstance.GUID, binding.GUID)).To(BeElementOf(1, 2))

		// This fails because it tries to generate a random number whose max value is lower than its min value
		By("running 'cf upgrade-service' and getting a failure")
		Expect(broker.UpgradeService(serviceInstance, endingVersion, testdrive.WithUpgradePreviousValues(domain.PreviousValues{PlanID: servicePlanGUID}))).To(MatchError("update failed with state: failed"))

		By("observing that the instance TF state file has been updated to the latest version, but not the binding")
		Expect(instanceTFStateVersion(serviceInstance.GUID)).To(Equal(endingVersion))
		Expect(bindingTFStateVersion(serviceInstance.GUID, binding.GUID)).To(Equal(startingVersion))

		By("observing the instance and binding outputs have not changed")
		Expect(instanceTFStateOutputValue(serviceInstance.GUID)).To(BeElementOf(1, 2))
		Expect(bindingTFStateOutputValue(serviceInstance.GUID, binding.GUID)).To(BeElementOf(1, 2))

		By("hacking the database so that the next upgrade will succeed")
		Expect(dbConn.Model(&models.ProvisionRequestDetails{}).Where("service_instance_id = ?", serviceInstance.GUID).Update("request_details", `{"max":8}`).Error).To(Succeed())

		By("running 'cf upgrade-service' again and succeeding")
		Expect(broker.UpgradeService(serviceInstance, endingVersion, testdrive.WithUpgradePreviousValues(domain.PreviousValues{PlanID: servicePlanGUID}))).To(Succeed())

		By("observing that the instance TF state file has been updated to the latest version for both the instance and the binding")
		Expect(instanceTFStateVersion(serviceInstance.GUID)).To(Equal(endingVersion))
		Expect(bindingTFStateVersion(serviceInstance.GUID, binding.GUID)).To(Equal(endingVersion))

		By("observing the instance and binding output are in the updated range")
		Expect(instanceTFStateOutputValue(serviceInstance.GUID)).To(BeElementOf(3, 4))
		Expect(bindingTFStateOutputValue(serviceInstance.GUID, binding.GUID)).To(BeElementOf(3, 4))
	})
})
