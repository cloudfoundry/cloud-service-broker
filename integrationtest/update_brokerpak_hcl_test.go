package integrationtest_test

import (
	"fmt"

	"github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"

	"github.com/cloudfoundry/cloud-service-broker/v3/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/v3/internal/testdrive"

	"github.com/cloudfoundry/cloud-service-broker/v3/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/workspace"
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

		serviceOfferingGUID       = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		servicePlanGUID           = "8b52a460-b246-11eb-a8f5-d349948e2480"
		tfWorkspaceIDQuery        = "id = ?"
		initialBrokerpak          = "update-brokerpak-hcl"
		updatedBrokerpak          = "update-brokerpak-hcl-updated"
		updatedBrokerpakWithError = "update-brokerpak-hcl-updated-with-error"
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

	beUpdatedTerraformHCL := SatisfyAll(
		ContainSubstring(updatedOutputHCL),
		Not(ContainSubstring(initialOutputHCL)),
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

	pushUpdatedBrokerpakWithError := func(updatesEnabled bool) {
		Expect(broker.Stop()).To(Succeed())
		must(packer.BuildBrokerpak(csb, fixtures(updatedBrokerpakWithError), packer.WithDirectory(brokerpak)))
		startBroker(updatesEnabled)
	}

	verifyUpdatedHCLInUse := func(err error) {
		Expect(err).To(HaveOccurred())
		Expect(err).To(matchUnexpectedStatusErrorWithSubstring(500, `An input variable with the name \"wrong_variable\" has not been declared`))
	}

	verifyDeprovisionUpdatedHCLInUse := func(err error) {
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`An input variable with the name "wrong_update_input" has not been declared`))
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

		It("uses the original HCL for service instance update", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(false)

			Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(updateParams))).To(Succeed())

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(haveInitialOutputs)
		})

		It("ignores extra parameters added in the update", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(false)

			Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(`{"update_input":"update output value","extra_input":"foo"}`))).To(Succeed())

			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(haveInitialOutputs)
		})

		It("uses the original HCL for service instance deprovision", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpakWithError(false) // if it errors then it is using the updated HCL

			Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(updateParams))).To(Succeed())

			Expect(broker.Deprovision(serviceInstance)).To(Succeed())
		})

		It("uses the original HCL for unbinding", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			serviceBinding := must(broker.CreateBinding(serviceInstance))
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, serviceBinding.GUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, serviceBinding.GUID)).To(haveInitialBindingOutputs)

			pushUpdatedBrokerpakWithError(false) // if it errors then it is using the updated HCL

			Expect(broker.DeleteBinding(serviceInstance, serviceBinding.GUID)).To(Succeed())
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

		It("uses extra parameters added in the update", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(true)

			Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(`{"update_input":"update output value","extra_input":" and extra parameter"}`))).To(Succeed())

			Expect(persistedTerraformState(serviceInstance.GUID, "")).To(ContainSubstring(`"value": "update output value and extra parameter"`))
		})

		It("uses the updated HCL for a deprovision", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))

			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpakWithError(true) // if it errors then it is using the updated HCL

			verifyDeprovisionUpdatedHCLInUse(broker.Deprovision(serviceInstance))
		})

		It("uses the updated HCL for unbind", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			serviceBinding := must(broker.CreateBinding(serviceInstance, testdrive.WithBindingParams(bindParams)))
			Expect(persistedTerraformModuleDefinition(serviceInstance.GUID, serviceBinding.GUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstance.GUID, serviceBinding.GUID)).To(haveInitialBindingOutputs)

			pushUpdatedBrokerpakWithError(true) // if it errors then it is using the updated HCL

			verifyUpdatedHCLInUse(broker.DeleteBinding(serviceInstance, serviceBinding.GUID))
		})
	})
})

func matchUnexpectedStatusErrorWithSubstring(code int, subs string) types.GomegaMatcher {
	return gstruct.PointTo(gstruct.MatchAllFields(gstruct.Fields{
		"StatusCode":   Equal(code),
		"ResponseBody": ContainSubstring(subs),
	}))
}
