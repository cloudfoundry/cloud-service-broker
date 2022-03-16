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

var _ = Context("Terraform012Invoker", func() {
	var fakeExecutor *wrapperfakes.FakeTerraformExecutor
	var fakeWorkspace *tffakes.FakeWorkspace
	var invoker tf.TerraformInvoker
	var expectedContext = context.TODO()
	var pluginDirectory = "plugindir"

	BeforeEach(func() {
		fakeExecutor = &wrapperfakes.FakeTerraformExecutor{}
		fakeWorkspace = &tffakes.FakeWorkspace{}
		invoker = tf.NewTerraform012Invoker(fakeExecutor, pluginDirectory)
	})

	Context("Apply", func() {
		It("initializes the workspace and applies", func() {
			Expect(
				invoker.Apply(expectedContext, fakeWorkspace),
			).To(Succeed())

			Expect(fakeWorkspace.ExecuteCallCount()).To(Equal(1))
			actualContext, actualExecutor, actualCommands := fakeWorkspace.ExecuteArgsForCall(0)
			Expect(actualContext).To(Equal(expectedContext))
			Expect(actualExecutor).To(Equal(fakeExecutor))
			Expect(actualCommands).To(Equal([]wrapper.TerraformCommand{
				wrapper.NewInit012Command(pluginDirectory),
				wrapper.ApplyCommand{},
			}))
		})
	})
})
