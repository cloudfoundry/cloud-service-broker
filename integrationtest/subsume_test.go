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

var _ = Describe("Subsume", func() {
	var (
		originalDir helper.Original
		testLab     *helper.TestLab
		session     *Session
	)

	BeforeEach(func() {
		originalDir = helper.OriginalDir()
		testLab = helper.NewTestLab(csb)
		testLab.BuildBrokerpak(string(originalDir), "fixtures", "brokerpak-for-subsume-cancel")
		session = testLab.StartBroker()
	})

	AfterEach(func() {
		session.Terminate()
		originalDir.Return()
	})

	It("can subsume a resource", func() {
		const serviceOfferingGUID = "547cad88-fa93-11eb-9f44-97feefe52547"
		const servicePlanGUID = "59624c68-fa93-11eb-9081-e79b0e1ab5ae"
		serviceInstanceGUID := uuid.New()
		provisionResponse := testLab.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), []byte(`{"value":"a97fd57a-fa94-11eb-8256-930255607a99"}`))
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted), string(provisionResponse.ResponseBody))

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

	It("cancels a subsume operation when a resource would be deleted", func() {
		// This test relies on a behaviour in the random string resource where it gets re-created after being imported
		const serviceOfferingGUID = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		const servicePlanGUID = "8b52a460-b246-11eb-a8f5-d349948e2481"
		serviceInstanceGUID := uuid.New()
		provisionResponse := testLab.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), []byte(`{"value":"thisisnotrandomatall"}`))
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted), string(provisionResponse.ResponseBody))

		var receiver domain.LastOperation
		Eventually(func() bool {
			lastOperationResponse := testLab.Client().LastOperation(serviceInstanceGUID, requestID())
			Expect(lastOperationResponse.Error).NotTo(HaveOccurred())
			Expect(lastOperationResponse.StatusCode).To(Equal(http.StatusOK))
			err := json.Unmarshal(lastOperationResponse.ResponseBody, &receiver)
			Expect(err).NotTo(HaveOccurred())
			Expect(receiver.State).NotTo(Equal("succeeded"))
			return receiver.State == "failed"
		}, time.Minute*2, time.Second*10).Should(BeTrue())
		Expect(receiver.Description).To(Equal("terraform plan shows that resources would be destroyed - cancelling subsume"))
	})
})
