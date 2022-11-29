package integrationtest_test

import (
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Preventing destroy during update", func() {
	const (
		serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		servicePlanGUID     = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
	)

	It("prevents a service instance being destroyed during an update", func() {
		brokerpak := must(packer.BuildBrokerpak(csb, fixtures("prevent-destroy-during-update")))
		broker := must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})

		By("provisioning with default length")
		serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))

		By("failing update when the resource would be deleted")
		updateResponse := broker.Client.Update(serviceInstance.GUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(`{"length":5}`), domain.PreviousValues{}, nil)
		Expect(updateResponse.Error).NotTo(HaveOccurred())
		Expect(updateResponse.StatusCode).To(Equal(http.StatusAccepted), string(updateResponse.ResponseBody))
		Expect(broker.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Failed))
		Expect(must(broker.LastOperation(serviceInstance.GUID)).Description).To(ContainSubstring("Error: Instance cannot be destroyed"))

		By("successfully deleting")
		Expect(broker.Deprovision(serviceInstance)).To(Succeed())
	})
})
