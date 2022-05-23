package integrationtest_test

import (
	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Multiple Updates to Properties", func() {
	const (
		serviceOfferingGUID = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		servicePlanGUID     = "8b52a460-b246-11eb-a8f5-d349948e2480"
	)

	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "multiple-update-properties")
		session = testHelper.StartBroker()
	})

	AfterEach(func() {
		session.Terminate()
	})

	// This test was added for issue https://www.pivotaltracker.com/story/show/178213626 where a parameter that was
	// updated would be reverted to the default value in subsequent updates
	It("persists updated parameters in subsequent updates", func() {
		By("provisioning with parameters")
		const provisionParams = `{"alpha_input":"foo","beta_input":"bar"}`
		serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID, provisionParams)

		By("checking that the parameter values are in a binding")
		_, bindingOutput := testHelper.CreateBinding(serviceInstance)
		Expect(bindingOutput).To(ContainSubstring(`"bind_output":"foo;bar"`))

		By("updating a parameter")
		const updateOneParams = `{"beta_input":"baz"}`
		testHelper.UpdateService(serviceInstance, updateOneParams)

		By("checking the value is updated in a binding")
		_, bindingOutput = testHelper.CreateBinding(serviceInstance)
		Expect(bindingOutput).To(ContainSubstring(`"bind_output":"foo;baz"`))

		By("updating another parameter")
		const updateTwoParams = `{"alpha_input":"quz"}`
		testHelper.UpdateService(serviceInstance, updateTwoParams)

		By("checking that both parameters remain updated in a binding")
		_, bindingOutput = testHelper.CreateBinding(serviceInstance)
		Expect(bindingOutput).To(ContainSubstring(`"bind_output":"quz;baz"`))
	})
})
