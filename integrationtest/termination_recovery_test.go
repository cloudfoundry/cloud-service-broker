package integrationtest_test

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/v2/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/testdrive"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/brokerapi/v11/domain"
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
	Describe("running csb on a VM", func() {
		BeforeEach(func() {
			brokerpak = must(packer.BuildBrokerpak(csb, fixtures("termination-recovery")))

			stdout = gbytes.NewBuffer()
			stderr = gbytes.NewBuffer()
			os.RemoveAll("/tmp/csb/")
			broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr)))

			DeferCleanup(func() {
				Expect(broker.Terminate()).To(Succeed())
				cleanup(brokerpak)
			})
		})
		It("can finish the in flight operation", func() {
			By("starting to provision")
			instanceGUID := uuid.NewString()
			response := broker.Client.Provision(instanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.NewString(), nil)

			Expect(response.Error).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusAccepted))
			Eventually(stdout, time.Second*5).Should(gbytes.Say(`tofu","apply","-auto-approve"`))
			By("gracefully stopping the broker")
			// Stop seems to be blocking, so run it in a routine so we can check that the broker actually rejects requests until it's fully stopped.
			go func() {
				Expect(broker.Stop()).To(Succeed())
			}()

			By("logging a message")
			Eventually(stdout).Should(gbytes.Say("received SIGTERM"))

			By("ensuring  that the broker rejects requests")
			Expect(broker.Client.LastOperation(instanceGUID, uuid.NewString()).Error).To(HaveOccurred())

			Eventually(stdout, time.Minute).Should(Say("draining csb"))
			Eventually(stdout, time.Minute).Should(Say("draining complete"))
			Eventually(stdout, time.Minute).ShouldNot(Say("shutdown error"))

			broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr)))

			By("checking that the resource finished successfully")
			response = broker.Client.LastOperation(instanceGUID, uuid.NewString())
			Expect(string(response.ResponseBody)).To(ContainSubstring(`{"state":"succeeded","description":"provision succeeded"}`))
			Expect(response.Error).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			By("ensuring SI can be successfully deleted")
			si := testdrive.ServiceInstance{GUID: instanceGUID, ServiceOfferingGUID: serviceOfferingGUID, ServicePlanGUID: servicePlanGUID}
			Expect(broker.Deprovision(si)).To(Succeed())
		})
	})
	Describe("running csb as a CF app", func() {
		BeforeEach(func() {
			brokerpak = must(packer.BuildBrokerpak(csb, fixtures("termination-recovery")))

			stdout = NewBuffer()
			stderr = NewBuffer()
			os.RemoveAll("/tmp/csb/")
			broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr), testdrive.WithEnv("CF_INSTANCE_GUID=dcfa061e-c0e3-4237-a805-734578347393")))

			DeferCleanup(func() {
				Expect(broker.Terminate()).To(Succeed())
				cleanup(brokerpak)
			})
		})

		It("can recover from a terminated create", func() {
			By("starting to provision")
			instanceGUID := uuid.NewString()
			response := broker.Client.Provision(instanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.NewString(), nil)
			Expect(response.Error).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusAccepted))

			By("terminating and restarting the broker")
			Expect(broker.Terminate()).To(Succeed())
			broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr), testdrive.WithEnv("CF_INSTANCE_GUID=dcfa061e-c0e3-4237-a805-734578347393")))

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
			response = broker.Client.Provision(instanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.NewString(), nil)
			Expect(response.Error).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusConflict))

			By("allowing the instance to be cleaned up")
			response = broker.Client.Deprovision(instanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.NewString())
			Expect(response.Error).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
		})

		It("can recover from a terminated update", func() {
			By("successfully provisioning a service instance")
			instance, err := broker.Provision(serviceOfferingGUID, servicePlanGUID)
			Expect(err).NotTo(HaveOccurred())

			By("starting to update")
			response := broker.Client.Update(instance.GUID, serviceOfferingGUID, servicePlanGUID, uuid.NewString(), nil, domain.PreviousValues{}, nil)
			Expect(response.Error).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusAccepted))

			By("terminating and restarting the broker")
			Expect(broker.Terminate()).To(Succeed())
			broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr), testdrive.WithEnv("CF_INSTANCE_GUID=dcfa061e-c0e3-4237-a805-734578347393")))

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
			response := broker.Client.Deprovision(instance.GUID, serviceOfferingGUID, servicePlanGUID, uuid.NewString())
			Expect(response.Error).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusAccepted))

			By("terminating and restarting the broker")
			Expect(broker.Terminate()).To(Succeed())
			broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr), testdrive.WithEnv("CF_INSTANCE_GUID=dcfa061e-c0e3-4237-a805-734578347393")))

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
			bindingGUID := uuid.NewString()
			go broker.CreateBinding(instance, testdrive.WithBindingGUID(bindingGUID))

			Eventually(stdout).Should(Say(fmt.Sprintf(`"cloud-service-broker.Binding".*"binding_id":"%s"`, bindingGUID)))

			By("terminating and restarting the broker")
			Expect(broker.Terminate()).To(Succeed())
			broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr), testdrive.WithEnv("CF_INSTANCE_GUID=dcfa061e-c0e3-4237-a805-734578347393")))

			By("allowing the operation to be restarted")
			_, err = broker.CreateBinding(instance, testdrive.WithBindingGUID(bindingGUID))
			Expect(err).NotTo(HaveOccurred())
		})

		It("can recover from a terminated unbind", func() {
			By("successfully provisioning a service instance and binding")
			instance, err := broker.Provision(serviceOfferingGUID, servicePlanGUID)
			Expect(err).NotTo(HaveOccurred())

			bindingGUID := uuid.NewString()
			_, err = broker.CreateBinding(instance, testdrive.WithBindingGUID(bindingGUID))
			Expect(err).NotTo(HaveOccurred())

			By("starting to unbind")
			go broker.DeleteBinding(instance, bindingGUID)

			Eventually(stdout).Should(Say(fmt.Sprintf(`"cloud-service-broker.Unbinding".*"binding_id":"%s"`, bindingGUID)))

			By("terminating and restarting the broker")
			Expect(broker.Terminate()).To(Succeed())
			broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(stdout, stderr), testdrive.WithEnv("CF_INSTANCE_GUID=dcfa061e-c0e3-4237-a805-734578347393")))

			By("allowing the operation to be restarted")
			Expect(broker.DeleteBinding(instance, bindingGUID)).To(Succeed())
		})
	})
})
