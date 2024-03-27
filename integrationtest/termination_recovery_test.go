package integrationtest_test

import (
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v10/domain"
)

var _ = Describe("Recovery From Broker Termination", func() {
	const (
		serviceOfferingGUID = "083f2884-eb7b-11ee-96c7-174e35671015"
		servicePlanGUID     = "0d953850-eb7b-11ee-bb2c-8ba95d780d82"
	)

	var (
		brokerpak string
		broker    *testdrive.Broker
		stdout    *Buffer
		stderr    *Buffer
	)

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("termination-recovery")))

		stdout = NewBuffer()
		stderr = NewBuffer()
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	It("can recover from a terminated create", func() {
		By("starting to provision")
		instanceGUID := uuid.New()
		response := broker.Client.Provision(instanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), nil)
		Expect(response.Error).NotTo(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusAccepted))

		By("terminating and restarting the broker")
		Expect(broker.Stop()).To(Succeed())
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr)))

		By("reporting that an operation failed")
		lastOperation, err := broker.LastOperation(instanceGUID)
		Expect(err).NotTo(HaveOccurred())
		Expect(lastOperation.Description).To(Equal("the broker restarted while the operation was in progress"))
		Expect(lastOperation.State).To(BeEquivalentTo("failed"))

		By("logging a message")
		ws := fmt.Sprintf(`"workspace_id":"tf:%s:"`, instanceGUID)
		Expect(string(stdout.Contents())).To(SatisfyAll(ContainSubstring("recover-in-progress-operations.mark-as-failed"), ContainSubstring(ws)))

		// OSBAPI requires that HTTP 409 (Conflict) is returned
		By("refusing to allow a duplicate instance")
		response = broker.Client.Provision(instanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), nil)
		Expect(response.Error).NotTo(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusConflict))

		By("allowing the instance to be cleaned up")
		response = broker.Client.Deprovision(instanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New())
		Expect(response.Error).NotTo(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
	})

	It("can recover from a terminated update", func() {
		By("successfully provisioning a service instance")
		instance, err := broker.Provision(serviceOfferingGUID, servicePlanGUID)
		Expect(err).NotTo(HaveOccurred())

		By("starting to update")
		response := broker.Client.Update(instance.GUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), nil, domain.PreviousValues{}, nil)
		Expect(response.Error).NotTo(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusAccepted))

		By("terminating and restarting the broker")
		Expect(broker.Stop()).To(Succeed())
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr)))

		By("reporting that an operation failed")
		lastOperation, err := broker.LastOperation(instance.GUID)
		Expect(err).NotTo(HaveOccurred())
		Expect(lastOperation.Description).To(Equal("the broker restarted while the operation was in progress"))
		Expect(lastOperation.State).To(BeEquivalentTo("failed"))

		By("logging a message")
		ws := fmt.Sprintf(`"workspace_id":"tf:%s:"`, instance.GUID)
		Expect(string(stdout.Contents())).To(SatisfyAll(ContainSubstring("recover-in-progress-operations.mark-as-failed"), ContainSubstring(ws)))

		By("allowing the operation to be restarted")
		Expect(broker.UpdateService(instance)).To(Succeed())
	})

	It("can recover from a terminated delete", func() {
		By("successfully provisioning a service instance")
		instance, err := broker.Provision(serviceOfferingGUID, servicePlanGUID)
		Expect(err).NotTo(HaveOccurred())

		By("starting to delete")
		response := broker.Client.Deprovision(instance.GUID, serviceOfferingGUID, servicePlanGUID, uuid.New())
		Expect(response.Error).NotTo(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusAccepted))

		By("terminating and restarting the broker")
		Expect(broker.Stop()).To(Succeed())
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr)))

		By("reporting that an operation failed")
		lastOperation, err := broker.LastOperation(instance.GUID)
		Expect(err).NotTo(HaveOccurred())
		Expect(lastOperation.Description).To(Equal("the broker restarted while the operation was in progress"))
		Expect(lastOperation.State).To(BeEquivalentTo("failed"))

		By("logging a message")
		ws := fmt.Sprintf(`"workspace_id":"tf:%s:"`, instance.GUID)
		Expect(string(stdout.Contents())).To(SatisfyAll(ContainSubstring("recover-in-progress-operations.mark-as-failed"), ContainSubstring(ws)))

		By("allowing the operation to be restarted")
		Expect(broker.Deprovision(instance)).To(Succeed())
	})

	It("can recover from a terminated bind", func() {
		By("successfully provisioning a service instance")
		instance, err := broker.Provision(serviceOfferingGUID, servicePlanGUID)
		Expect(err).NotTo(HaveOccurred())

		By("starting to bind")
		bindingGUID := uuid.New()
		go broker.CreateBinding(instance, testdrive.WithBindingGUID(bindingGUID))

		Eventually(stdout).Should(Say(fmt.Sprintf(`"cloud-service-broker.Binding".*"binding_id":"%s"`, bindingGUID)))

		By("terminating and restarting the broker")
		Expect(broker.Stop()).To(Succeed())
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr)))

		By("allowing the operation to be restarted")
		_, err = broker.CreateBinding(instance, testdrive.WithBindingGUID(bindingGUID))
		Expect(err).NotTo(HaveOccurred())
	})

	It("can recover from a terminated unbind", func() {
		By("successfully provisioning a service instance and binding")
		instance, err := broker.Provision(serviceOfferingGUID, servicePlanGUID)
		Expect(err).NotTo(HaveOccurred())

		bindingGUID := uuid.New()
		_, err = broker.CreateBinding(instance, testdrive.WithBindingGUID(bindingGUID))
		Expect(err).NotTo(HaveOccurred())

		By("starting to unbind")
		go broker.DeleteBinding(instance, bindingGUID)

		Eventually(stdout).Should(Say(fmt.Sprintf(`"cloud-service-broker.Unbinding".*"binding_id":"%s"`, bindingGUID)))

		By("terminating and restarting the broker")
		Expect(broker.Stop()).To(Succeed())
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr)))

		By("allowing the operation to be restarted")
		Expect(broker.DeleteBinding(instance, bindingGUID)).To(Succeed())
	})
})
