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
		session.Terminate().Wait()
	})

	// This test was added for issue https://www.pivotaltracker.com/story/show/178213626 where a parameter that was
	// updated would be reverted to the default value in subsequent updates
	It("persists updated parameters in subsequent updates", func() {
		By("provisioning with a parameter")
		serviceInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID, `{"beta_input":"foo"}`)

		By("checking that the parameter value, and the default value are in a binding")
		_, bindingOutput := testHelper.CreateBinding(serviceInstance)
		Expect(bindingOutput).To(ContainSubstring(`"bind_output":"default_alpha;foo"`))

		By("updating both parameters")
		testHelper.UpdateService(serviceInstance, `{"alpha_input":"foo","beta_input":"bar"}`)

		By("checking that the parameter values are in a binding")
		_, bindingOutput = testHelper.CreateBinding(serviceInstance)
		Expect(bindingOutput).To(ContainSubstring(`"bind_output":"foo;bar"`))

		By("updating just one parameter")
		testHelper.UpdateService(serviceInstance, `{"beta_input":"baz"}`)

		By("checking that just one value is updated in a binding")
		_, bindingOutput = testHelper.CreateBinding(serviceInstance)
		Expect(bindingOutput).To(ContainSubstring(`"bind_output":"foo;baz"`))

		By("updating another parameter")
		testHelper.UpdateService(serviceInstance, `{"alpha_input":"quz"}`)

		By("checking that both parameters remain updated in a binding")
		_, bindingOutput = testHelper.CreateBinding(serviceInstance)
		Expect(bindingOutput).To(ContainSubstring(`"bind_output":"quz;baz"`))

		By("unsetting parameters")
		testHelper.UpdateService(serviceInstance, `{"alpha_input":"","beta_input":null}`)
		_, bindingOutput = testHelper.CreateBinding(serviceInstance)
		Expect(bindingOutput).To(ContainSubstring(`"bind_output":";is_null"`))
	})
})
