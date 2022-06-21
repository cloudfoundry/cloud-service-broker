package upgrader_test

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/upgrader"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/upgrader/upgraderfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Upgrade", func() {
	const fakePlanGUID = "test-plan-guid"
	const fakeBrokerName = "fake-broker-name"

	var (
		fakeCCAPI                                           *upgraderfakes.FakeCCAPI
		fakePlan                                            ccapi.Plan
		fakeInstance1, fakeInstance2, fakeInstanceNoUpgrade ccapi.ServiceInstance
	)

	BeforeEach(func() {
		fakePlan = ccapi.Plan{
			GUID:                   fakePlanGUID,
			MaintenanceInfoVersion: "test-maintenance-info",
		}
		fakeInstance1 = ccapi.ServiceInstance{
			GUID:             "fake-instance-guid-1",
			PlanGUID:         fakePlanGUID,
			UpgradeAvailable: true,
		}
		fakeInstance2 = ccapi.ServiceInstance{
			GUID:             "fake-instance-guid-2",
			PlanGUID:         fakePlanGUID,
			UpgradeAvailable: true,
		}

		fakeInstanceNoUpgrade = ccapi.ServiceInstance{
			GUID:             "fake-instance-no-upgrade-GUID",
			PlanGUID:         fakePlanGUID,
			UpgradeAvailable: false,
		}

		fakeCCAPI = &upgraderfakes.FakeCCAPI{}
		fakeCCAPI.GetServicePlansReturns([]ccapi.Plan{fakePlan}, nil)
		fakeCCAPI.GetServiceInstancesReturns([]ccapi.ServiceInstance{fakeInstance1, fakeInstance2, fakeInstanceNoUpgrade}, nil)
	})

	It("upgrades a service instance", func() {
		err := upgrader.Upgrade(fakeCCAPI, fakeBrokerName)
		Expect(err).NotTo(HaveOccurred())

		By("getting the service plans")
		Expect(fakeCCAPI.GetServicePlansCallCount()).To(Equal(1))
		Expect(fakeCCAPI.GetServicePlansArgsForCall(0)).To(Equal(fakeBrokerName))

		By("getting the service instances")
		Expect(fakeCCAPI.GetServiceInstancesCallCount()).To(Equal(1))
		Expect(fakeCCAPI.GetServiceInstancesArgsForCall(0)).To(Equal([]string{fakePlanGUID}))

		By("calling upgrade on each upgradeable instance")
		Expect(fakeCCAPI.UpgradeServiceInstanceCallCount()).Should(Equal(2))
		guids := []string{fakeCCAPI.UpgradeServiceInstanceArgsForCall(0), fakeCCAPI.UpgradeServiceInstanceArgsForCall(1)}
		Expect(guids).To(ConsistOf("fake-instance-guid-1", "fake-instance-guid-2"))
	})

	When("getting service plans fails", func() {
		It("returns the error", func() {
			fakeCCAPI.GetServicePlansReturns(nil, fmt.Errorf("plan-error"))

			err := upgrader.Upgrade(fakeCCAPI, fakeBrokerName)
			Expect(err).To(MatchError("plan-error"))
		})
	})

	When("getting service instances fails", func() {
		It("returns the error", func() {
			fakeCCAPI.GetServiceInstancesReturns(nil, fmt.Errorf("instance-error"))

			err := upgrader.Upgrade(fakeCCAPI, fakeBrokerName)
			Expect(err).To(MatchError("instance-error"))
		})
	})
})
