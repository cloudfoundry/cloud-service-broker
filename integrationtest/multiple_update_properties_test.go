package integrationtest_test

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Multiple Updates to Properties", func() {
	const (
		serviceOfferingGUID = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		servicePlanGUID     = "8b52a460-b246-11eb-a8f5-d349948e2480"
	)

	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	BeforeEach(func() {
		brokerpak = must(testdrive.BuildBrokerpak(csb, fixtures("multiple-update-properties")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))
	})

	AfterEach(func() {
		Expect(broker.Stop()).To(Succeed())
		cleanup(brokerpak)
	})

	// This test was added for issue https://www.pivotaltracker.com/story/show/178213626 where a parameter that was
	// updated would be reverted to the default value in subsequent updates
	It("persists updated parameters in subsequent updates", func() {
		By("provisioning with a parameter")
		serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(`{"beta_input":"foo"}`)))

		By("checking that the parameter value, and the default value are in a binding")
		binding := must(broker.CreateBinding(serviceInstance))
		Expect(binding.Body).To(ContainSubstring(`"bind_output":"default_alpha;foo"`))

		By("updating both parameters")
		Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(`{"alpha_input":"foo","beta_input":"bar"}`))).To(Succeed())

		By("checking that the parameter values are in a binding")
		binding = must(broker.CreateBinding(serviceInstance))
		Expect(binding.Body).To(ContainSubstring(`"bind_output":"foo;bar"`))

		By("updating just one parameter")
		Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(`{"beta_input":"baz"}`))).To(Succeed())

		By("checking that just one value is updated in a binding")
		binding = must(broker.CreateBinding(serviceInstance))
		Expect(binding.Body).To(ContainSubstring(`"bind_output":"foo;baz"`))

		By("updating another parameter")
		Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(`{"alpha_input":"quz"}`))).To(Succeed())

		By("checking that both parameters remain updated in a binding")
		binding = must(broker.CreateBinding(serviceInstance))
		Expect(binding.Body).To(ContainSubstring(`"bind_output":"quz;baz"`))

		By("unsetting parameters")
		Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(`{"alpha_input":"","beta_input":null}`))).To(Succeed())
		binding = must(broker.CreateBinding(serviceInstance))
		Expect(binding.Body).To(ContainSubstring(`"bind_output":";is_null"`))
	})
})
