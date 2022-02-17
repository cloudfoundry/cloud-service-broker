package integrationtest_test

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Multiple Updates to Properties", func() {
	const (
		serviceOfferingGUID = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		servicePlanGUID     = "8b52a460-b246-11eb-a8f5-d349948e2480"
	)

	var (
		testHelper          *helper.TestHelper
		session             *Session
		serviceInstanceGUID string
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "brokerpak-for-multiple-updates")
		session = testHelper.StartBroker()
	})

	AfterEach(func() {
		session.Terminate()
		testHelper.Restore()
	})

	checkBindingOutput := func(expected string) {
		bindResponse := testHelper.Client().Bind(serviceInstanceGUID, uuid.New(), serviceOfferingGUID, servicePlanGUID, requestID(), nil)
		ExpectWithOffset(1, bindResponse.Error).NotTo(HaveOccurred())
		ExpectWithOffset(1, string(bindResponse.ResponseBody)).To(ContainSubstring(expected))
	}

	waitForCompletion := func() {
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

	// This test was added for issue https://www.pivotaltracker.com/story/show/178213626 where a parameter that was
	// updated would be reverted to the default value in subsequent updates
	It("persists updated parameters in subsequent updates", func() {
		By("provisioning with parameters")
		const provisionParams = `{"alpha_input":"foo","beta_input":"bar"}`
		serviceInstanceGUID = uuid.New()
		provisionResponse := testHelper.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(provisionParams))
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForCompletion()

		By("checking that the parameter values are in a binding")
		checkBindingOutput(`"bind_output":"foo;bar"`)

		By("updating a parameter")
		const updateOneParams = `{"beta_input":"baz"}`
		updateOneResponse := testHelper.Client().Update(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(updateOneParams))
		Expect(updateOneResponse.Error).NotTo(HaveOccurred())
		Expect(updateOneResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForCompletion()

		By("checking the value is updated in a binding")
		checkBindingOutput(`"bind_output":"foo;baz"`)

		By("updating another parameter")
		const updateTwoParams = `{"alpha_input":"quz"}`
		updateTwoResponse := testHelper.Client().Update(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(updateTwoParams))
		Expect(updateTwoResponse.Error).NotTo(HaveOccurred())
		Expect(updateTwoResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForCompletion()

		By("checking that both parameters remain updated in a binding")
		checkBindingOutput(`"bind_output":"quz;baz"`)
	})
})
