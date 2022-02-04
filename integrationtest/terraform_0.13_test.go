package integrationtest_test

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cloudfoundry-incubator/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	_ "gorm.io/driver/sqlite"
)

var _ = Describe("Terraform 0.13", func() {
	var (
		originalDir helper.Original
		testLab     *helper.TestLab
		session     *Session
	)

	BeforeEach(func() {
		originalDir = helper.OriginalDir()
		testLab = helper.NewTestLab(csb)
		testLab.BuildBrokerpak(string(originalDir), "fixtures", "brokerpak-terraform-0.13")
		session = testLab.StartBroker()
	})

	AfterEach(func() {
		session.Terminate()
		originalDir.Return()
	})

	It("can provision using this Terraform version", func() {
		const serviceOfferingGUID = "29d4119f-2e88-4e85-8c40-7360f3d9c695"
		const servicePlanGUID = "056c052c-e7b0-4c8e-8d6b-18616e06a7ac"
		serviceInstanceGUID := uuid.New()
		provisionResponse := testLab.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), nil)
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))

		Eventually(func() bool {
			lastOperationResponse := testLab.Client().LastOperation(serviceInstanceGUID, requestID())
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
