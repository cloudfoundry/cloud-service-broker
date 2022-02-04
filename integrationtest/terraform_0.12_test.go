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

var _ = Describe("Terraform 0.12", func() {
	var (
		originalDir helper.Original
		testLab     *helper.TestLab
		session     *Session
	)

	BeforeEach(func() {
		originalDir = helper.OriginalDir()
		testLab = helper.NewTestLab(csb)
		testLab.BuildBrokerpak(string(originalDir), "fixtures", "brokerpak-terraform-0.12")

		session = testLab.StartBroker()
	})

	AfterEach(func() {
		session.Terminate()
		originalDir.Return()
	})

	It("can provision using this Terraform version", func() {
		const serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		const servicePlanGUID = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
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
