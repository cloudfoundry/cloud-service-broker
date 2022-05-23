package integrationtest_test

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Update Brokerpak HCL", func() {
	const (
		provisionParams             = `{"provision_input":"bar"}`
		bindParams                  = `{"bind_input":"quz"}`
		updateParams                = `{"update_input": "update output value"}`
		updateOutputStateKey        = `"update_output"`
		updatedUpdateOutputStateKey = `"update_output_updated"`
		updatedOutputHCL            = `output update_output_updated { value = "${var.update_input == null ? "empty" : var.update_input }${var.extra_input == null ? "empty" : var.extra_input}" }`
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
		testHelper *helper.TestHelper
		session    *Session
	)

	findRecord := func(dest interface{}, query, guid string) {
		result := testHelper.DBConn().Where(query, guid).First(dest)
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

	startBroker := func(updatesEnabled bool) *Session {
		return testHelper.StartBroker(fmt.Sprintf("BROKERPAK_UPDATES_ENABLED=%t", updatesEnabled))
	}

	persistedTerraformModuleDefinition := func(serviceInstanceGUID, serviceBindingGUID string) string {
		return persistedTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID).Modules[0].Definitions["main"]
	}

	persistedTerraformState := func(serviceInstanceGUID, serviceBindingGUID string) string {
		return string(persistedTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID).State)
	}

	pushUpdatedBrokerpak := func(updatesEnabled bool) {
		session.Terminate().Wait()
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", updatedBrokerpak)
		session = startBroker(updatesEnabled)
	}

	BeforeEach(func() {
		testHelper = helper.New(csb)

		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", initialBrokerpak)
	})

	AfterEach(func() {
		session.Terminate()
	})

	When("brokerpak updates are disabled", func() {
		BeforeEach(func() {
			session = startBroker(false)
		})

		It("uses the original HCL for service instances operations", func() {
			serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID, provisionParams)
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(false)

			testHelper.UpdateService(serviceInstance, updateParams)

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(haveInitialOutputs)

			testHelper.Deprovision(serviceInstance)
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(haveEmptyState)
		})

		It("uses the original HCL for binding operations", func() {
			serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID, provisionParams)
			serviceBindingGUID, _ := testHelper.CreateBinding(serviceInstance)
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, serviceBindingGUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, serviceBindingGUID)).To(haveInitialBindingOutputs)

			pushUpdatedBrokerpak(false)

			testHelper.DeleteBinding(serviceInstance, serviceBindingGUID)

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, serviceBindingGUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, serviceBindingGUID)).To(haveEmptyState)
		})

		It("ignores extra parameters added in the update", func() {
			serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID, provisionParams)
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(false)

			testHelper.UpdateService(serviceInstance, `{"update_input":"update output value","extra_input":"foo"}`)

			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(haveInitialOutputs)
		})
	})

	When("brokerpak updates are enabled", func() {
		BeforeEach(func() {
			session = startBroker(true)
		})

		It("uses the updated HCL for an update", func() {
			serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID, provisionParams)
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(true)

			testHelper.UpdateService(serviceInstance, updateParams)

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beUpdatedTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(haveUpdatedOutputs)
		})

		It("uses the updated HCL for a deprovision", func() {
			serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID, provisionParams)

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(true)

			testHelper.Deprovision(serviceInstance)

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beUpdatedTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(haveEmptyState)
		})

		It("uses the updated HCL for unbind", func() {
			serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID, provisionParams)
			serviceBindingGUID, _ := testHelper.CreateBinding(serviceInstance, bindParams)
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, serviceBindingGUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, serviceBindingGUID)).To(haveInitialBindingOutputs)

			pushUpdatedBrokerpak(true)

			testHelper.DeleteBinding(serviceInstance, serviceBindingGUID)

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, serviceBindingGUID)).To(beUpdatedBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, serviceBindingGUID)).To(haveEmptyState)
		})

		It("uses extra parameters added in the update", func() {
			serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID, provisionParams)

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(true)

			testHelper.UpdateService(serviceInstance, `{"update_input":"update output value","extra_input":" and extra parameter"}`)

			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(ContainSubstring(`"value": "update output value and extra parameter"`))
		})
	})
})
