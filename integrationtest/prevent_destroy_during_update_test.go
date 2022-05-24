package integrationtest_test

import (
	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Preventing destroy during update", Ordered, func() {
	const (
		serviceOfferingGUID = "df2c1512-3013-11ec-8704-2fbfa9c8a802"
		servicePlanGUID     = "e59773ce-3013-11ec-9bbb-9376b4f72d14"
	)

	var (
		testHelper      *helper.TestHelper
		serviceInstance helper.ServiceInstance
		session         *Session
	)

	BeforeAll(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "prevent-destroy-during-update")

		session = testHelper.StartBroker()
		DeferCleanup(func() {
			session.Terminate().Wait()
		})
	})

	It("provisions with default length", func() {
		serviceInstance = testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
	})

	It("fails update when the resource would be deleted", func() {
		testHelper.Client().Update(serviceInstance.GUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(`{"length":5}`))
		Expect(testHelper.LastOperationFinalState(serviceInstance.GUID)).To(Equal(domain.Failed))
		Expect(testHelper.LastOperation(serviceInstance.GUID).Description).To(ContainSubstring("Error: Instance cannot be destroyed"))
	})

	It("can be successfully deleted", func() {
		testHelper.Deprovision(serviceInstance)
	})
})
