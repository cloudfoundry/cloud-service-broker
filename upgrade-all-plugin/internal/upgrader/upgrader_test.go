package upgrader_test

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/upgrader"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/upgrader/upgraderfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Upgrade", func() {
	const fakePlanGUID = "test-plan-guid"
	const fakeInstanceGUID = "test-instance-guid"
	const fakeBrokerName = "fake-broker-name"

	var (
		fakeCCAPI                           *upgraderfakes.FakeCCAPI
		fakePlan                            upgrader.Plan
		fakeInstance, fakeInstanceNoUpgrade upgrader.ServiceInstance
	)

	BeforeEach(func() {
		fakePlan = upgrader.Plan{
			GUID:                   fakePlanGUID,
			MaintenanceInfoVersion: "test-maintenance-info",
		}

		fakeInstance = upgrader.ServiceInstance{
			GUID:             fakeInstanceGUID,
			PlanGUID:         fakePlanGUID,
			UpgradeAvailable: true,
		}

		fakeInstanceNoUpgrade = upgrader.ServiceInstance{
			GUID:             "fake-instance-no-upgrade-GUID",
			PlanGUID:         fakePlanGUID,
			UpgradeAvailable: false,
		}

		fakeCCAPI = &upgraderfakes.FakeCCAPI{}

		fakeCCAPI.GetServicePlansReturns([]upgrader.Plan{fakePlan}, nil)
		fakeCCAPI.GetServiceInstancesReturns([]upgrader.ServiceInstance{fakeInstance, fakeInstance, fakeInstanceNoUpgrade}, nil)
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
		Eventually(fakeCCAPI.UpgradeServiceInstanceCallCount()).Should(Equal(2))
		Eventually(fakeCCAPI.UpgradeServiceInstanceArgsForCall(0)).Should(Equal(fakeInstanceGUID))
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

	// Gets the service plans
	// Gets service instances
	// It upgrades the service instances with upgradeAvailable
})
