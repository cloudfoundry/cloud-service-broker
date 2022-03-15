package integrationtest_test

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Terraform 0.13", func() {
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

	It("can provision using this Terraform version", func() {
		const serviceOfferingGUID = "29d4119f-2e88-4e85-8c40-7360f3d9c695"
		const servicePlanGUID = "056c052c-e7b0-4c8e-8d6b-18616e06a7ac"
		serviceInstanceGUID := uuid.New()
		provisionResponse := testHelper.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), nil)
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))

		Eventually(func() bool {
			lastOperationResponse := testHelper.Client().LastOperation(serviceInstanceGUID, requestID())
			Expect(lastOperationResponse.Error).NotTo(HaveOccurred())
			Expect(lastOperationResponse.StatusCode).To(Equal(http.StatusOK))
			var receiver domain.LastOperation
			err := json.Unmarshal(lastOperationResponse.ResponseBody, &receiver)
			Expect(err).NotTo(HaveOccurred())
			Expect(receiver.State).NotTo(Equal("failed"))
			return receiver.State == "succeeded"
		}, time.Minute*2, time.Second*10).Should(BeTrue())
	})
})
