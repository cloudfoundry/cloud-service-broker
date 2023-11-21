package invoker_test

import (
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor/executorfakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/invoker"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/gomega"
)

var _ = Context("TerraformInvokerFactory", func() {
	var fakeBuilder *executorfakes.FakeExecutorBuilder
	var fakeExecutor *executorfakes.FakeTerraformExecutor
	var invokerFactory invoker.TerraformInvokerBuilder
	var expectedTerraformPluginDir = "a_dir"
	var expectedProviderRenames = map[string]string{"from": "to"}

	BeforeEach(func() {
		fakeBuilder = &executorfakes.FakeExecutorBuilder{}
		fakeExecutor = &executorfakes.FakeTerraformExecutor{}
		fakeBuilder.VersionedExecutorReturns(fakeExecutor)

		invokerFactory = invoker.NewTerraformInvokerFactory(fakeBuilder, expectedTerraformPluginDir, expectedProviderRenames)
	})
	Context("0.12", func() {
		It("should return invoker for 0.12, for terraform version 0.12.0", func() {
			Expect(
				invokerFactory.VersionedTerraformInvoker(newVersion("0.12.0")),
			).To(Equal(invoker.NewTerraform012Invoker(fakeExecutor, expectedTerraformPluginDir)))
			Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
			Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("0.12.0")))
		})

		It("should return invoker for 0.12, for terraform version 0.12.1", func() {
			Expect(
				invokerFactory.VersionedTerraformInvoker(newVersion("0.12.1")),
			).To(Equal(invoker.NewTerraform012Invoker(fakeExecutor, expectedTerraformPluginDir)))
			Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
			Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("0.12.1")))
		})
	})
	Context("0.13+", func() {
		It("should return default invoker, for terraform version 0.13.1", func() {
			Expect(
				invokerFactory.VersionedTerraformInvoker(newVersion("0.13.1")),
			).To(Equal(invoker.NewTerraformDefaultInvoker(fakeExecutor, expectedTerraformPluginDir, expectedProviderRenames)))
			Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
			Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("0.13.1")))
		})

		It("should return default invoker, for terraform version 1.0.4", func() {
			Expect(
				invokerFactory.VersionedTerraformInvoker(newVersion("1.0.4")),
			).To(Equal(invoker.NewTerraformDefaultInvoker(fakeExecutor, expectedTerraformPluginDir, expectedProviderRenames)))
			Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
			Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("1.0.4")))
		})
	})
})

func newVersion(v string) *version.Version {
	return version.Must(version.NewVersion(v))
}
