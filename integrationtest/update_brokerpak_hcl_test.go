package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/integrationtest/helper"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf/wrapper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
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
		tfWorkspaceIdQuery  = "id = ?"
		initialBrokerpak    = "brokerpak-with-fake-provider"
		updatedBrokerpak    = "brokerpak-with-fake-provider-updated"
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

	persistedTerraformWorkspace := func(serviceInstanceGUID, serviceBindingGUID string) *wrapper.TerraformWorkspace {
		record := models.TerraformDeployment{}
		findRecord(&record, tfWorkspaceIdQuery, fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, serviceBindingGUID))
		ws, _ := wrapper.DeserializeWorkspace(string(record.Workspace))
		return ws
	}

	createBinding := func(serviceInstanceGUID, serviceBindingGUID string) {
		bindResponse := testHelper.Client().Bind(serviceInstanceGUID, serviceBindingGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(bindParams))
		Expect(bindResponse.Error).NotTo(HaveOccurred())
		Expect(bindResponse.StatusCode).To(Equal(http.StatusCreated))
	}

	deleteBinding := func(serviceInstanceGUID, serviceBindingGUID string) {
		unbindResponse := testHelper.Client().Unbind(serviceInstanceGUID, serviceBindingGUID, serviceOfferingGUID, servicePlanGUID, requestID())
		Expect(unbindResponse.Error).NotTo(HaveOccurred())
		Expect(unbindResponse.StatusCode).To(Equal(http.StatusOK))
	}

	waitForAsyncRequest := func(serviceInstanceGUID string) {
		Eventually(func() bool {
			lastOperationResponse := testHelper.Client().LastOperation(serviceInstanceGUID, requestID())
			Expect(lastOperationResponse.Error).NotTo(HaveOccurred())
			Expect(lastOperationResponse.StatusCode).To(Equal(http.StatusOK))
			var receiver domain.LastOperation
			err := json.Unmarshal(lastOperationResponse.ResponseBody, &receiver)
			Expect(err).NotTo(HaveOccurred())
			Expect(receiver.State).NotTo(Equal("failed"))
			return receiver.State == "succeeded"
		}, time.Minute*2, time.Second*10).Should(BeTrue())
	}

	updateServiceInstance := func(serviceInstanceGUID, updateParams string) {
		updateResponse := testHelper.Client().Update(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(updateParams))
		Expect(updateResponse.Error).NotTo(HaveOccurred())
		Expect(updateResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForAsyncRequest(serviceInstanceGUID)
	}

	provisionServiceInstance := func(serviceInstanceGUID string) {
		provisionResponse := testHelper.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), []byte(provisionParams))
		ExpectWithOffset(1, provisionResponse.Error).NotTo(HaveOccurred())
		ExpectWithOffset(1, provisionResponse.StatusCode).To(Equal(http.StatusAccepted), string(provisionResponse.ResponseBody))

		waitForAsyncRequest(serviceInstanceGUID)
	}

	deprovisionServiceInstance := func(serviceInstanceGUID string) {
		deprovisionResponse := testHelper.Client().Deprovision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID())
		Expect(deprovisionResponse.Error).NotTo(HaveOccurred())
		Expect(deprovisionResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForAsyncRequest(serviceInstanceGUID)
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

	haveEmtpyState := SatisfyAll(
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
		session.Terminate()
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", updatedBrokerpak)
		session = startBroker(updatesEnabled)
	}

	BeforeEach(func() {
		testHelper = helper.New(csb)

		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", initialBrokerpak)
	})

	AfterEach(func() {
		session.Terminate()
		testHelper.Restore()
	})

	When("brokerpak updates are disabled", func() {
		var (
			serviceInstanceGUID string
			serviceBindingGUID  string
		)

		BeforeEach(func() {
			serviceInstanceGUID = uuid.New()
			serviceBindingGUID = uuid.New()

			session = startBroker(false)
			provisionServiceInstance(serviceInstanceGUID)
		})

		It("uses the original HCL for service instances operations", func() {
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(false)

			updateServiceInstance(serviceInstanceGUID, updateParams)

			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beInitialTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, "")).To(haveInitialOutputs)

			deprovisionServiceInstance(serviceInstanceGUID)
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beInitialTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, "")).To(haveEmtpyState)
		})

		It("uses the original HCL for binding operations", func() {
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, serviceBindingGUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, serviceBindingGUID)).To(haveInitialBindingOutputs)

			pushUpdatedBrokerpak(false)

			deleteBinding(serviceInstanceGUID, serviceBindingGUID)

			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, serviceBindingGUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, serviceBindingGUID)).To(haveEmtpyState)
		})

		It("ignores extra parameters added in the update", func() {
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(false)

			updateServiceInstance(serviceInstanceGUID, `{"update_input":"update output value","extra_input":"foo"}`)

			Expect(persistedTerraformState(serviceInstanceGUID, "")).To(haveInitialOutputs)
		})
	})

	When("brokerpak updates are enabled", func() {
		var (
			serviceInstanceGUID string
			serviceBindingGUID  string
		)

		BeforeEach(func() {
			serviceInstanceGUID = uuid.New()
			serviceBindingGUID = uuid.New()

			session = startBroker(true)
			provisionServiceInstance(serviceInstanceGUID)
		})

		It("uses the updated HCL for an update", func() {
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(true)

			updateServiceInstance(serviceInstanceGUID, updateParams)

			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beUpdatedTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, "")).To(haveUpdatedOutputs)
		})

		It("uses the updated HCL for a deprovision", func() {
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(true)

			deprovisionServiceInstance(serviceInstanceGUID)

			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beUpdatedTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, "")).To(haveEmtpyState)
		})

		It("uses the updated HCL for unbind", func() {
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, serviceBindingGUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, serviceBindingGUID)).To(haveInitialBindingOutputs)

			pushUpdatedBrokerpak(true)

			deleteBinding(serviceInstanceGUID, serviceBindingGUID)

			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, serviceBindingGUID)).To(beUpdatedBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, serviceBindingGUID)).To(haveEmtpyState)
		})

		It("uses extra parameters added in the update", func() {
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(true)

			updateServiceInstance(serviceInstanceGUID, `{"update_input":"update output value","extra_input":" and extra parameter"}`)

			Expect(persistedTerraformState(serviceInstanceGUID, "")).To(ContainSubstring(`"value": "update output value and extra parameter"`))
		})
	})
})
