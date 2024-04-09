package invoker_test

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/command"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/invoker"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/executor/executorfakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace/workspacefakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("TerraformDefaultInvoker", func() {
	var fakeExecutor *executorfakes.FakeTerraformExecutor
	var fakeWorkspace *workspacefakes.FakeWorkspace
	var invokerUnderTest invoker.TerraformInvoker
	var expectedContext = context.TODO()
	var pluginDirectory = "plugindir"
	var providerRenames = map[string]string{
		"old_provider_1": "new_provider_1",
	}

	BeforeEach(func() {
		fakeExecutor = &executorfakes.FakeTerraformExecutor{}
		fakeWorkspace = &workspacefakes.FakeWorkspace{}
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
					command.NewInit(pluginDirectory),
					command.NewApply(),
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
					command.NewRenameProvider("old_provider_1", "new_provider_1"),
					command.NewInit(pluginDirectory),
					command.NewApply(),
				}))
			})
		})
	})

	Context("Destroy", func() {
		Context("has no renames", func() {
			BeforeEach(func() {
				invokerUnderTest = invoker.NewTerraformDefaultInvoker(fakeExecutor, pluginDirectory, nil)
			})
			It("initializes the workspace and applies", func() {
				invokerUnderTest.Destroy(expectedContext, fakeWorkspace)

				Expect(fakeWorkspace.ExecuteCallCount()).To(Equal(1))
				actualContext, actualExecutor, actualCommands := fakeWorkspace.ExecuteArgsForCall(0)
				Expect(actualContext).To(Equal(expectedContext))
				Expect(actualExecutor).To(Equal(fakeExecutor))
				Expect(actualCommands).To(Equal([]command.TerraformCommand{
					command.NewInit(pluginDirectory),
					command.NewDestroy(),
				}))
			})
		})
		Context("has renames", func() {
			It("renames providers before, initializing the workspace and applies", func() {
				invokerUnderTest.Destroy(expectedContext, fakeWorkspace)

				Expect(fakeWorkspace.ExecuteCallCount()).To(Equal(1))
				actualContext, actualExecutor, actualCommands := fakeWorkspace.ExecuteArgsForCall(0)
				Expect(actualContext).To(Equal(expectedContext))
				Expect(actualExecutor).To(Equal(fakeExecutor))
				Expect(actualCommands).To(Equal([]command.TerraformCommand{
					command.NewRenameProvider("old_provider_1", "new_provider_1"),
					command.NewInit(pluginDirectory),
					command.NewDestroy(),
				}))
			})
		})
	})

	Context("Show", func() {
		Context("has no renames", func() {
			BeforeEach(func() {
				invokerUnderTest = invoker.NewTerraformDefaultInvoker(fakeExecutor, pluginDirectory, nil)
			})
			It("initializes the workspace and show", func() {
				invokerUnderTest.Show(expectedContext, fakeWorkspace)

				Expect(fakeWorkspace.ExecuteCallCount()).To(Equal(1))
				actualContext, actualExecutor, actualCommands := fakeWorkspace.ExecuteArgsForCall(0)
				Expect(actualContext).To(Equal(expectedContext))
				Expect(actualExecutor).To(Equal(fakeExecutor))
				Expect(actualCommands).To(Equal([]command.TerraformCommand{
					command.NewInit(pluginDirectory),
					command.NewShow(),
				}))
			})
		})
		Context("has renames", func() {
			It("renames providers before, initializing the workspace and applies", func() {
				invokerUnderTest.Show(expectedContext, fakeWorkspace)

				Expect(fakeWorkspace.ExecuteCallCount()).To(Equal(1))
				actualContext, actualExecutor, actualCommands := fakeWorkspace.ExecuteArgsForCall(0)
				Expect(actualContext).To(Equal(expectedContext))
				Expect(actualExecutor).To(Equal(fakeExecutor))
				Expect(actualCommands).To(Equal([]command.TerraformCommand{
					command.NewRenameProvider("old_provider_1", "new_provider_1"),
					command.NewInit(pluginDirectory),
					command.NewShow(),
				}))
			})
		})
	})
})
