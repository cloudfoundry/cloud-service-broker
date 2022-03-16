package invoker_test

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/command"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/invoker"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor/executorfakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/tffakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Terpkg/providers/tf/invokerraformDefaultInvoker", func() {
	var fakeExecutor *executorfakes.FakeTerraformExecutor
	var fakeWorkspace *tffakes.FakeWorkspace
	var invokerUnderTest invoker.TerraformInvoker
	var expectedContext = context.TODO()
	var pluginDirectory = "plugindir"
	var providerRenames = map[string]string{
		"old_provider_1": "new_provider_1",
	}

	BeforeEach(func() {
		fakeExecutor = &executorfakes.FakeTerraformExecutor{}
		fakeWorkspace = &tffakes.FakeWorkspace{}
		invokerUnderTest = invoker.NewTerraformDefaultInvoker(fakeExecutor, pluginDirectory, providerRenames)
	})

	Context("Apply", func() {
		Context("workspace has no state", func() {
			BeforeEach(func() {
				fakeWorkspace.HasStateReturns(false)
			})
			It("initializes the workspace and applies", func() {
				invokerUnderTest.Apply(expectedContext, fakeWorkspace)

				Expect(fakeWorkspace.ExecuteCallCount()).To(Equal(1))
				actualContext, actualExecutor, actualCommands := fakeWorkspace.ExecuteArgsForCall(0)
				Expect(actualContext).To(Equal(expectedContext))
				Expect(actualExecutor).To(Equal(fakeExecutor))
				Expect(actualCommands).To(Equal([]command.TerraformCommand{
					command.NewInitCommand(pluginDirectory),
					command.ApplyCommand{},
				}))
			})
		})
		Context("workspace has existing state", func() {
			BeforeEach(func() {
				fakeWorkspace.HasStateReturns(true)
			})
			It("renames providers before, initializing the workspace and applies", func() {
				invokerUnderTest.Apply(expectedContext, fakeWorkspace)

				Expect(fakeWorkspace.ExecuteCallCount()).To(Equal(1))
				actualContext, actualExecutor, actualCommands := fakeWorkspace.ExecuteArgsForCall(0)
				Expect(actualContext).To(Equal(expectedContext))
				Expect(actualExecutor).To(Equal(fakeExecutor))
				Expect(actualCommands).To(Equal([]command.TerraformCommand{
					command.NewRenameProviderCommand("old_provider_1", "new_provider_1"),
					command.NewInitCommand(pluginDirectory),
					command.ApplyCommand{},
				}))
			})
		})
	})
})
