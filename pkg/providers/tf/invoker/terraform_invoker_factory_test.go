package invoker_test

import (
	"github.com/cloudfoundry/cloud-service-broker/pkg/featureflags"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/executor/executorfakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/invoker"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
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

	Context("1.2.0+", func() {
		Context("with provider renames disabled", func() {
			BeforeEach(func() {
				viper.Set(string(featureflags.DisableTfUpgradeProviderRenames), true)

			})
			AfterEach(func() {
				viper.Reset()
			})

			It("should return an invoker with renames, for terraform version 1.1.9", func() {
				Expect(
					invokerFactory.VersionedTerraformInvoker(newVersion("1.1.9")),
				).To(Equal(invoker.NewTerraformDefaultInvoker(fakeExecutor, expectedTerraformPluginDir, expectedProviderRenames)))
				Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
				Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("1.1.9")))
			})

			It("should return an invoker with no renames, for terraform version 1.2.0", func() {
				Expect(
					invokerFactory.VersionedTerraformInvoker(newVersion("1.2.0")),
				).To(Equal(invoker.NewTerraformDefaultInvoker(fakeExecutor, expectedTerraformPluginDir, map[string]string{})))
				Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
				Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("1.2.0")))
			})

			It("should return an invoker with no renames, for terraform version 1.4.2", func() {
				Expect(
					invokerFactory.VersionedTerraformInvoker(newVersion("1.4.2")),
				).To(Equal(invoker.NewTerraformDefaultInvoker(fakeExecutor, expectedTerraformPluginDir, map[string]string{})))
				Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
				Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("1.4.2")))
			})
		})

		Context("with provider renames enabled", func() {
			It("should return an invoker that performs renames, for 1.2.0", func() {
				Expect(
					invokerFactory.VersionedTerraformInvoker(newVersion("1.2.0")),
				).To(Equal(invoker.NewTerraformDefaultInvoker(fakeExecutor, expectedTerraformPluginDir, expectedProviderRenames)))
				Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
				Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("1.2.0")))
			})

			It("should return an invoker that performs renames, for 1.4.2", func() {
				Expect(
					invokerFactory.VersionedTerraformInvoker(newVersion("1.4.2")),
				).To(Equal(invoker.NewTerraformDefaultInvoker(fakeExecutor, expectedTerraformPluginDir, expectedProviderRenames)))
				Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
				Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("1.4.2")))
			})

			It("should return an invoker with renames, for terraform version 1.1.9", func() {
				Expect(
					invokerFactory.VersionedTerraformInvoker(newVersion("1.1.9")),
				).To(Equal(invoker.NewTerraformDefaultInvoker(fakeExecutor, expectedTerraformPluginDir, expectedProviderRenames)))
				Expect(fakeBuilder.VersionedExecutorCallCount()).To(Equal(1))
				Expect(fakeBuilder.VersionedExecutorArgsForCall(0)).To(Equal(newVersion("1.1.9")))
			})
		})
	})
})

func newVersion(v string) *version.Version {
	return version.Must(version.NewVersion(v))
}
