package tf_test

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/tffakes"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PollInstance", func() {
	It("returns gets operation status", func() {
		fakeDeploymentManager := &tffakes.FakeDeploymentManagerInterface{}
		fakeInvokerBuilder := &tffakes.FakeTerraformInvokerBuilder{}
		fakeLogger := utils.NewLogger("test")

		fakeDeploymentManager.OperationStatusReturns(true, "LO message", "update", nil)
		provider := tf.NewTerraformProvider(executor.TFBinariesContext{}, fakeInvokerBuilder, fakeLogger, tf.TfServiceDefinitionV1{}, fakeDeploymentManager)

		finished, message, operationType, err := provider.PollInstance(context.TODO(), "instance-guid")

		Expect(err).NotTo(HaveOccurred())
		Expect(finished).To(BeTrue())
		Expect(operationType).To(Equal("update"))
		Expect(message).To(Equal("LO message"))

		Expect(fakeDeploymentManager.OperationStatusCallCount()).To(Equal(1))
		Expect(fakeDeploymentManager.OperationStatusArgsForCall(0)).To(Equal("tf:instance-guid:"))
	})
})
