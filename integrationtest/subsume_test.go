package integrationtest_test

import (
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Subsume", func() {
	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "subsume")
		session = testHelper.StartBroker()
	})

	AfterEach(func() {
		session.Terminate()
	})

	It("can subsume a resource", func() {
		const serviceOfferingGUID = "547cad88-fa93-11eb-9f44-97feefe52547"
		const servicePlanGUID = "59624c68-fa93-11eb-9081-e79b0e1ab5ae"
		testHelper.Provision(serviceOfferingGUID, servicePlanGUID, `{"value":"a97fd57a-fa94-11eb-8256-930255607a99"}`)
	})

	It("cancels a subsume operation when a resource would be deleted", func() {
		// This test relies on a behaviour in the random string resource where it gets re-created after being imported
		const serviceOfferingGUID = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		const servicePlanGUID = "8b52a460-b246-11eb-a8f5-d349948e2481"
		serviceInstanceGUID := uuid.New()
		provisionResponse := testHelper.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), []byte(`{"value":"thisisnotrandomatall"}`))
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted), string(provisionResponse.ResponseBody))
		Expect(testHelper.LastOperationFinalState(serviceInstanceGUID)).To(Equal(domain.Failed))
		Expect(testHelper.LastOperation(serviceInstanceGUID).Description).To(Equal("provision failed: terraform plan shows that resources would be destroyed - cancelling subsume"))
	})
})
