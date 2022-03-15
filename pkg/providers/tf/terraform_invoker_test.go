package tf_test

import (
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper/wrapperfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("TerraformInvokerFactory", func() {
	var fakeBuilder *wrapperfakes.FakeExecutorBuilder
	var fakeExecutor *wrapperfakes.FakeTerraformExecutor
	var invokerFactory tf.TerraformInvokerBuilder
	var expectedTerraformPluginDir = "a_dir"
	var expectedProviderRenames = map[string]string{"from": "to"}

	BeforeEach(func() {
		fakeBuilder = &wrapperfakes.FakeExecutorBuilder{}
		fakeExecutor = &wrapperfakes.FakeTerraformExecutor{}
		fakeBuilder.VersionedExecutorReturns(fakeExecutor)

		invokerFactory = tf.NewTerraformInvokerFactory(fakeBuilder, expectedTerraformPluginDir, expectedProviderRenames)
	})
	Context("0.12", func() {
		It("should return invoker for 0.12, for terraform version 0.12.0", func() {
			Expect(
				invokerFactory.VersionedTerraformInvoker(newVersion("0.12.0")),
			).To(Equal(tf.NewTerraform012Invoker(fakeExecutor, expectedTerraformPluginDir)))
			Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
			Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("0.12.0")))
		})

		It("should return invoker for 0.12, for terraform version 0.12.1", func() {
			Expect(
				invokerFactory.VersionedTerraformInvoker(newVersion("0.12.1")),
			).To(Equal(tf.NewTerraform012Invoker(fakeExecutor, expectedTerraformPluginDir)))
			Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
			Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("0.12.1")))
		})
	})
	Context("0.13+", func() {
		It("should return default invoker, for terraform version 0.13.1", func() {
			Expect(
				invokerFactory.VersionedTerraformInvoker(newVersion("0.13.1")),
			).To(Equal(tf.NewTerraformDefaultInvoker(fakeExecutor, expectedTerraformPluginDir, expectedProviderRenames)))
			Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
			Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("0.13.1")))
		})

		It("should return default invoker, for terraform version 1.0.4", func() {
			Expect(
				invokerFactory.VersionedTerraformInvoker(newVersion("1.0.4")),
			).To(Equal(tf.NewTerraformDefaultInvoker(fakeExecutor, expectedTerraformPluginDir, expectedProviderRenames)))
			Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
			Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("1.0.4")))
		})
	})
})
