package invoker_test

import (
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/executor/executorfakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/invoker"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
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
	It("should return default invoker, for tofu version 1.6.0", func() {
		Expect(
			invokerFactory.VersionedTerraformInvoker(newVersion("1.6.0")),
		).To(Equal(invoker.NewTerraformDefaultInvoker(fakeExecutor, expectedTerraformPluginDir, expectedProviderRenames)))
		Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
		Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("1.6.0")))
	})

	It("should return default invoker, for tofu version 1.6.2", func() {
		Expect(
			invokerFactory.VersionedTerraformInvoker(newVersion("1.6.2")),
		).To(Equal(invoker.NewTerraformDefaultInvoker(fakeExecutor, expectedTerraformPluginDir, expectedProviderRenames)))
		Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
		Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("1.6.2")))
	})
})

func newVersion(v string) *version.Version {
	return version.Must(version.NewVersion(v))
}
