package integrationtest_test

import (
	"net/http"
	"time"

	"github.com/pivotal-cf/brokerapi/v8/domain"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
)

var _ = Describe("Terraform", func() {
	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "brokerpak-terraform-0.13")
		session = testHelper.StartBroker()
	})

	AfterEach(func() {
		session.Terminate()
		testHelper.Restore()
	})

	It("can provision when provider is renamed", func() {
		const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
		serviceInstanceGUID := uuid.New()
		provisionResponse := testHelper.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), nil)
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))

		Eventually(pollLastOperation(testHelper, serviceInstanceGUID), time.Minute*2, lastOperationPollingFrequency).ShouldNot(Equal(domain.InProgress))
		Expect(pollLastOperation(testHelper, serviceInstanceGUID)()).To(Equal(domain.Succeeded))

		session.Terminate().Wait()

		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "brokerpak-terraform-0.13-with-renamed-provider")
		session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=true", "BROKERPAK_UPDATES_ENABLED=true")

		By("running 'cf update-service'")
		updateResponse := testHelper.Client().Update(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(`{"alpha_input":"quz"}`))
		Expect(updateResponse.Error).NotTo(HaveOccurred())
		Expect(updateResponse.StatusCode).To(Equal(http.StatusAccepted))

		Eventually(pollLastOperation(testHelper, serviceInstanceGUID), time.Minute*2, lastOperationPollingFrequency).ShouldNot(Equal(domain.InProgress))
		Expect(pollLastOperation(testHelper, serviceInstanceGUID)()).To(Equal(domain.Succeeded))
	})

	It("can delete instance when provider is renamed", func() {
		const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
		serviceInstanceGUID := uuid.New()
		provisionResponse := testHelper.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), nil)
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))

		Eventually(pollLastOperation(testHelper, serviceInstanceGUID), time.Minute*2, lastOperationPollingFrequency).ShouldNot(Equal(domain.InProgress))
		Expect(pollLastOperation(testHelper, serviceInstanceGUID)()).To(Equal(domain.Succeeded))

		session.Terminate().Wait()

		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "brokerpak-terraform-0.13-with-renamed-provider")
		session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=true", "BROKERPAK_UPDATES_ENABLED=true")

		By("running 'cf delete-service'")
		deprovision := testHelper.Client().Deprovision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID())
		Expect(deprovision.Error).NotTo(HaveOccurred())
		Expect(deprovision.StatusCode).To(Equal(http.StatusAccepted))

		Eventually(pollLastOperation(testHelper, serviceInstanceGUID), time.Minute*2, lastOperationPollingFrequency).ShouldNot(Equal(domain.InProgress))
	})

	It("can delete binding when provider is renamed", func() {
		const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
		serviceInstanceGUID := uuid.New()
		provisionResponse := testHelper.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), nil)
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))

		Eventually(pollLastOperation(testHelper, serviceInstanceGUID), time.Minute*2, lastOperationPollingFrequency).ShouldNot(Equal(domain.InProgress))
		Expect(pollLastOperation(testHelper, serviceInstanceGUID)()).To(Equal(domain.Succeeded))

		bindingGUID := uuid.New()
		bindingResponse := testHelper.Client().Bind(serviceInstanceGUID, bindingGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), nil)
		Expect(bindingResponse.Error).NotTo(HaveOccurred())
		Expect(bindingResponse.StatusCode).To(Equal(http.StatusCreated))

		session.Terminate().Wait()

		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "brokerpak-terraform-0.13-with-renamed-provider")
		session = testHelper.StartBroker("TERRAFORM_UPGRADES_ENABLED=true", "BROKERPAK_UPDATES_ENABLED=true")

		By("running 'cf delete-binding'")
		updateResponse := testHelper.Client().Unbind(serviceInstanceGUID, bindingGUID, serviceOfferingGUID, servicePlanGUID, requestID())
		Expect(updateResponse.Error).NotTo(HaveOccurred())
		Expect(updateResponse.StatusCode).To(Equal(http.StatusOK))
	})
})
