package integrationtest_test

import (
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/v2/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/testdrive"

	"code.cloudfoundry.org/brokerapi/v13/domain"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Subsume", func() {
	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("subsume")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	It("can subsume a resource", func() {
		const serviceOfferingGUID = "547cad88-fa93-11eb-9f44-97feefe52547"
		const servicePlanGUID = "59624c68-fa93-11eb-9081-e79b0e1ab5ae"
		_, err := broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(`{"value":"a97fd57a-fa94-11eb-8256-930255607a99"}`))
		Expect(err).NotTo(HaveOccurred())
	})

	It("cancels a subsume operation when a resource would be deleted", func() {
		// This test relies on a behaviour in the random string resource where it gets re-created after being imported
		const serviceOfferingGUID = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		const servicePlanGUID = "8b52a460-b246-11eb-a8f5-d349948e2481"
		serviceInstanceGUID := uuid.NewString()
		provisionResponse := broker.Client.Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.NewString(), []byte(`{"value":"thisisnotrandomatall"}`))
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted), string(provisionResponse.ResponseBody))
		Expect(broker.LastOperationFinalState(serviceInstanceGUID)).To(Equal(domain.Failed))
		Expect(must(broker.LastOperation(serviceInstanceGUID)).Description).To(Equal("provision failed: tofu plan shows that resources would be destroyed - cancelling subsume"))
	})
})
