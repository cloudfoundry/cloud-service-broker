package integrationtest_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v11/domain"

	"github.com/cloudfoundry/cloud-service-broker/v3/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v3/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/v3/internal/testdrive"
)

var _ = Describe("Global labels propagation", Label("global-labels"), func() {
	const serviceOfferingGUID = "cc2005ea-2bb3-46e6-a505-248d43bffcc4"
	const servicePlanGUID = "d0af9bdf-3795-4e96-857a-db0050463d72"

	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	startBroker := func(startBrokerOptions testdrive.StartBrokerOption) {
		broker = must(testdrive.StartBroker(
			csb, brokerpak, database,
			testdrive.WithOutputs(GinkgoWriter, GinkgoWriter),
			startBrokerOptions,
		))
	}

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("global-labels")))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	getOutputsFromTFState := func(serviceInstanceGUID string) map[string]any {
		var tfDeploymentReceiver models.TerraformDeployment
		var workspaceReceiver struct {
			State []byte `json:"tfstate"`
		}
		var stateReceiver struct {
			Outputs map[string]any `json:"outputs"`
		}

		Expect(dbConn.Where("id = ?", fmt.Sprintf("tf:%s:", serviceInstanceGUID)).First(&tfDeploymentReceiver).Error).To(Succeed())

		Expect(json.Unmarshal(tfDeploymentReceiver.Workspace, &workspaceReceiver)).NotTo(HaveOccurred())
		Expect(json.Unmarshal(workspaceReceiver.State, &stateReceiver)).NotTo(HaveOccurred())
		return stateReceiver.Outputs
	}

	Describe("Provision and update a service", func() {
		When("global labels are set", func() {
			It("propagates the labels", func() {
				By("provisioning a service instance")
				brokerpakConfigEnv := `GSB_BROKERPAK_CONFIG={"global_labels":[{"key":  "key1", "value":  "value1"},{"key":  "key2", "value":  "value2"}]}`
				startBroker(
					testdrive.WithEnv(brokerpakConfigEnv),
				)

				serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
				expectedLabels := fmt.Sprintf(`{"key1":"value1","key2":"value2","pcf-instance-id":"%s","pcf-organization-guid":"","pcf-space-guid":""}`, serviceInstance.GUID)
				outputs := getOutputsFromTFState(serviceInstance.GUID)

				tfMap := outputs["labels"].(map[string]any) // TF creates a map with two labels: type and value.
				Expect(tfMap["value"]).To(Equal(expectedLabels))
				Expect(broker.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Succeeded))
				Expect(must(broker.LastOperation(serviceInstance.GUID)).Description).To(Equal("provision succeeded"))

				By("updating a service instance")
				err := broker.UpdateService(serviceInstance)

				Expect(err).To(Succeed())

				outputs = getOutputsFromTFState(serviceInstance.GUID)

				tfMap = outputs["labels"].(map[string]any) // TF creates a map with two labels: type and value.
				Expect(tfMap["value"]).To(Equal(expectedLabels))
				Expect(broker.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Succeeded))
				Expect(must(broker.LastOperation(serviceInstance.GUID)).Description).To(Equal("update succeeded"))
			})
		})

		When("global labels are not set", func() {
			It("only uses the default labels", func() {
				By("provisioning a service instance")
				brokerpakConfigEnv := `GSB_BROKERPAK_CONFIG={"another_config":[]}`
				startBroker(
					testdrive.WithEnv(brokerpakConfigEnv),
				)

				serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
				outputs := getOutputsFromTFState(serviceInstance.GUID)

				expectedLabels := fmt.Sprintf(`{"pcf-instance-id":"%s","pcf-organization-guid":"","pcf-space-guid":""}`, serviceInstance.GUID)
				tfMap := outputs["labels"].(map[string]any) // TF creates a map with two labels: type and value.
				Expect(tfMap["value"]).To(Equal(expectedLabels))
				Expect(broker.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Succeeded))
				Expect(must(broker.LastOperation(serviceInstance.GUID)).Description).To(Equal("provision succeeded"))

				By("updating a service instance")
				err := broker.UpdateService(serviceInstance)

				Expect(err).To(Succeed())

				outputs = getOutputsFromTFState(serviceInstance.GUID)

				tfMap = outputs["labels"].(map[string]any) // TF creates a map with two labels: type and value.
				Expect(tfMap["value"]).To(Equal(expectedLabels))
				Expect(broker.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Succeeded))
				Expect(must(broker.LastOperation(serviceInstance.GUID)).Description).To(Equal("update succeeded"))
			})
		})

		When("no global configuration is passed", func() {
			It("only uses the default labels", func() {
				By("provisioning a service instance")
				env := `FAKE_ENV={}`
				startBroker(
					testdrive.WithEnv(env),
				)

				serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
				outputs := getOutputsFromTFState(serviceInstance.GUID)

				expectedLabels := fmt.Sprintf(`{"pcf-instance-id":"%s","pcf-organization-guid":"","pcf-space-guid":""}`, serviceInstance.GUID)
				tfMap := outputs["labels"].(map[string]any) // TF creates a map with two labels: type and value.
				Expect(tfMap["value"]).To(Equal(expectedLabels))
				Expect(broker.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Succeeded))
				Expect(must(broker.LastOperation(serviceInstance.GUID)).Description).To(Equal("provision succeeded"))

				By("updating a service instance")
				err := broker.UpdateService(serviceInstance)

				Expect(err).To(Succeed())

				outputs = getOutputsFromTFState(serviceInstance.GUID)

				tfMap = outputs["labels"].(map[string]any) // TF creates a map with two labels: type and value.
				Expect(tfMap["value"]).To(Equal(expectedLabels))
				Expect(broker.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Succeeded))
				Expect(must(broker.LastOperation(serviceInstance.GUID)).Description).To(Equal("update succeeded"))
			})
		})
	})
})
