package integrationtest_test

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Update Brokerpak HCL", func() {
	const (
		provisionParams             = `{"provision_input":"bar"}`
		bindParams                  = `{"bind_input":"quz"}`
		updateParams                = `{"update_input": "update output value"}`
		updateOutputStateKey        = `"update_output"`
		updatedUpdateOutputStateKey = `"update_output_updated"`
		updatedOutputHCL            = `output update_output_updated { value = "${var.update_input == null ? "empty" : var.update_input}${var.extra_input == null ? "empty" : var.extra_input}" }`
		initialOutputHCL            = `output update_output { value = "${var.update_input}" }`
		initialBindingOutputHCL     = `bind_output { value = "${var.provision_output} and bind output value" }`
		updatedBindingOutputHCL     = `bind_output_updated { value = "${var.provision_output} and bind output value" }`
		bindingOutputStateKey       = `"bind_output"`

		serviceOfferingGUID = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		servicePlanGUID     = "8b52a460-b246-11eb-a8f5-d349948e2480"
		tfWorkspaceIDQuery  = "id = ?"
		initialBrokerpak    = "update-brokerpak-hcl"
		updatedBrokerpak    = "update-brokerpak-hcl-updated"
	)

	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	findRecord := func(dest any, query, guid string) {
		result := dbConn.Where(query, guid).First(dest)
		ExpectWithOffset(3, result.Error).NotTo(HaveOccurred())
		ExpectWithOffset(3, result.RowsAffected).To(Equal(int64(1)))
	}

	persistedTerraformWorkspace := func(serviceInstanceGUID, serviceBindingGUID string) *workspace.TerraformWorkspace {
		record := models.TerraformDeployment{}
		findRecord(&record, tfWorkspaceIDQuery, fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, serviceBindingGUID))
		ws, _ := workspace.DeserializeWorkspace(record.Workspace)
		return ws
	}

	beInitialTerraformHCL := SatisfyAll(
		ContainSubstring(initialOutputHCL),
		Not(ContainSubstring(updatedOutputHCL)),
	)

	beInitialBindingTerraformHCL := SatisfyAll(
		ContainSubstring(initialBindingOutputHCL),
		Not(ContainSubstring(updatedBindingOutputHCL)),
	)

	haveInitialOutputs := SatisfyAll(
		ContainSubstring(updateOutputStateKey),
		Not(ContainSubstring(updatedUpdateOutputStateKey)),
	)

	haveInitialBindingOutputs := ContainSubstring(bindingOutputStateKey)

	haveEmptyState := SatisfyAll(
		ContainSubstring(`"outputs": {}`),
		ContainSubstring(`"resources": []`),
	)

	beUpdatedTerraformHCL := SatisfyAll(
		ContainSubstring(updatedOutputHCL),
		Not(ContainSubstring(initialOutputHCL)),
	)

	beUpdatedBindingTerraformHCL := SatisfyAll(
		ContainSubstring(updatedBindingOutputHCL),
		Not(ContainSubstring(initialBindingOutputHCL)),
	)

	haveUpdatedOutputs := SatisfyAll(
		ContainSubstring(updatedUpdateOutputStateKey),
		Not(ContainSubstring(updateOutputStateKey)),
	)

	startBroker := func(updatesEnabled bool) {
		broker = must(testdrive.StartBroker(
			csb, brokerpak, database,
			testdrive.WithOutputs(GinkgoWriter, GinkgoWriter),
			testdrive.WithEnv(fmt.Sprintf("BROKERPAK_UPDATES_ENABLED=%t", updatesEnabled)),
		))
	}

	persistedTerraformModuleDefinition := func(serviceInstanceGUID, serviceBindingGUID string) string {
		return persistedTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID).Modules[0].Definitions["main"]
	}

	persistedTerraformState := func(serviceInstanceGUID, serviceBindingGUID string) string {
		return string(persistedTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID).State)
	}

	pushUpdatedBrokerpak := func(updatesEnabled bool) {
		Expect(broker.Stop()).To(Succeed())
		must(packer.BuildBrokerpak(csb, fixtures(updatedBrokerpak), packer.WithDirectory(brokerpak)))
		startBroker(updatesEnabled)
	}

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures(initialBrokerpak)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	When("brokerpak updates are disabled", func() {
		BeforeEach(func() {
			startBroker(false)
		})

		It("uses the original HCL for service instances operations", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(false)

			Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(updateParams))).To(Succeed())

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(haveInitialOutputs)

			Expect(broker.Deprovision(serviceInstance)).To(Succeed())
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(haveEmptyState)
		})

		It("uses the original HCL for binding operations", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			serviceBinding := must(broker.CreateBinding(serviceInstance))
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, serviceBinding.GUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, serviceBinding.GUID)).To(haveInitialBindingOutputs)

			pushUpdatedBrokerpak(false)

			Expect(broker.DeleteBinding(serviceInstance, serviceBinding.GUID)).To(Succeed())

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, serviceBinding.GUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, serviceBinding.GUID)).To(haveEmptyState)
		})

		It("ignores extra parameters added in the update", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(false)

			Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(`{"update_input":"update output value","extra_input":"foo"}`))).To(Succeed())

			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(haveInitialOutputs)
		})
	})

	When("brokerpak updates are enabled", func() {
		BeforeEach(func() {
			startBroker(true)
		})

		It("uses the updated HCL for an update", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(true)

			Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(updateParams))).To(Succeed())

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beUpdatedTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(haveUpdatedOutputs)
		})

		It("uses the updated HCL for a deprovision", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(true)

			Expect(broker.Deprovision(serviceInstance)).To(Succeed())

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beUpdatedTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(haveEmptyState)
		})

		It("uses the updated HCL for unbind", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			serviceBinding := must(broker.CreateBinding(serviceInstance, testdrive.WithBindingParams(bindParams)))
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, serviceBinding.GUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, serviceBinding.GUID)).To(haveInitialBindingOutputs)

			pushUpdatedBrokerpak(true)

			Expect(broker.DeleteBinding(serviceInstance, serviceBinding.GUID)).To(Succeed())

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, serviceBinding.GUID)).To(beUpdatedBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, serviceBinding.GUID)).To(haveEmptyState)
		})

		It("uses extra parameters added in the update", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(true)

			Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(`{"update_input":"update output value","extra_input":" and extra parameter"}`))).To(Succeed())

			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(ContainSubstring(`"value": "update output value and extra parameter"`))
		})
	})
})
