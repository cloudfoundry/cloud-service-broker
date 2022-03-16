package tf_test

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper/wrapperfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("TerraformDefaultInvoker", func() {
	var fakeExecutor *wrapperfakes.FakeTerraformExecutor
	var fakeWorkspace *tffakes.FakeWorkspace
	var invoker tf.TerraformInvoker
	var expectedContext = context.TODO()
	var pluginDirectory = "plugindir"
	var providerRenames = map[string]string{
		"old_provider_1": "new_provider_1",
	}

	BeforeEach(func() {
		fakeExecutor = &wrapperfakes.FakeTerraformExecutor{}
		fakeWorkspace = &tffakes.FakeWorkspace{}
		invoker = tf.NewTerraformDefaultInvoker(fakeExecutor, pluginDirectory, providerRenames)
	})

	Context("Apply", func() {
		Context("workspace has no state", func() {
			BeforeEach(func() {
				fakeWorkspace.HasStateReturns(false)
			})
			It("initializes the workspace and applies", func() {
				invoker.Apply(expectedContext, fakeWorkspace)

				Expect(fakeWorkspace.ExecuteCallCount()).To(Equal(1))
				actualContext, actualExecutor, actualCommands := fakeWorkspace.ExecuteArgsForCall(0)
				Expect(actualContext).To(Equal(expectedContext))
				Expect(actualExecutor).To(Equal(fakeExecutor))
				Expect(actualCommands).To(Equal([]wrapper.TerraformCommand{
					wrapper.NewInitCommand(pluginDirectory),
					wrapper.ApplyCommand{},
				}))
			})
		})
		Context("workspace has existing state", func() {
			BeforeEach(func() {
				fakeWorkspace.HasStateReturns(true)
			})
			It("renames providers before, initializing the workspace and applies", func() {
				invoker.Apply(expectedContext, fakeWorkspace)

				Expect(fakeWorkspace.ExecuteCallCount()).To(Equal(1))
				actualContext, actualExecutor, actualCommands := fakeWorkspace.ExecuteArgsForCall(0)
				Expect(actualContext).To(Equal(expectedContext))
				Expect(actualExecutor).To(Equal(fakeExecutor))
				Expect(actualCommands).To(Equal([]wrapper.TerraformCommand{
					wrapper.NewRenameProviderCommand("old_provider_1", "new_provider_1"),
					wrapper.NewInitCommand(pluginDirectory),
					wrapper.ApplyCommand{},
				}))
			})
		})
	})
})
