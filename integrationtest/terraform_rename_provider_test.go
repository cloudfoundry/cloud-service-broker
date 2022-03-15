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

		session.Terminate()
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "brokerpak-terraform-0.13-with-renamed-provider")

		session = testHelper.StartBroker()

		By("running 'cf update-service'")
		updateResponse := testHelper.Client().Update(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(`{"alpha_input":"quz"}`))
		Expect(updateResponse.Error).NotTo(HaveOccurred())
		Expect(updateResponse.StatusCode).To(Equal(http.StatusAccepted))

		Eventually(pollLastOperation(testHelper, serviceInstanceGUID), time.Minute*2, lastOperationPollingFrequency).ShouldNot(Equal(domain.InProgress))
		Expect(pollLastOperation(testHelper, serviceInstanceGUID)()).To(Equal(domain.Succeeded))
	})
})
