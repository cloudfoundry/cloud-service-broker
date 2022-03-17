package invoker_test

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor/executorfakes"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/command"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/invoker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/tffakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Terraform012Invoker", func() {
	var fakeExecutor *executorfakes.FakeTerraformExecutor
	var fakeWorkspace *tffakes.FakeWorkspace
	var invokerUnderTest invoker.TerraformInvoker
	var expectedContext = context.TODO()
	var pluginDirectory = "plugindir"

	BeforeEach(func() {
		fakeExecutor = &executorfakes.FakeTerraformExecutor{}
		fakeWorkspace = &tffakes.FakeWorkspace{}
		invokerUnderTest = invoker.NewTerraform012Invoker(fakeExecutor, pluginDirectory)
	})

	Context("Apply", func() {
		It("initializes the workspace and applies", func() {
			Expect(
				invokerUnderTest.Apply(expectedContext, fakeWorkspace),
			).To(Succeed())

			Expect(fakeWorkspace.ExecuteCallCount()).To(Equal(1))
			actualContext, actualExecutor, actualCommands := fakeWorkspace.ExecuteArgsForCall(0)
			Expect(actualContext).To(Equal(expectedContext))
			Expect(actualExecutor).To(Equal(fakeExecutor))
			Expect(actualCommands).To(Equal([]command.TerraformCommand{
				command.NewInit012Command(pluginDirectory),
				command.Apply{},
			}))
		})
	})
})
