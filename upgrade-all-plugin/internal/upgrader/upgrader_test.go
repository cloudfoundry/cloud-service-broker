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
	const (
		fakePlanGUID   = "test-plan-guid"
		fakeBrokerName = "fake-broker-name"
	)

	var (
		fakeCFClient          *upgraderfakes.FakeCFClient
		fakePlan              ccapi.Plan
		fakeInstance1         ccapi.ServiceInstance
		fakeInstance2         ccapi.ServiceInstance
		fakeInstanceNoUpgrade ccapi.ServiceInstance
		fakeServiceInstances  []ccapi.ServiceInstance
		fakeLog               *upgraderfakes.FakeLogger
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

		fakeServiceInstances = []ccapi.ServiceInstance{fakeInstance1, fakeInstance2, fakeInstanceNoUpgrade}

		fakeCFClient = &upgraderfakes.FakeCFClient{}
		fakeCFClient.GetServicePlansReturns([]ccapi.Plan{fakePlan}, nil)
		fakeCFClient.GetServiceInstancesReturns(fakeServiceInstances, nil)

		fakeLog = &upgraderfakes.FakeLogger{}
	})

	It("upgrades a service instance", func() {
		err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 5, fakeLog)
		Expect(err).NotTo(HaveOccurred())

		By("getting the service plans")
		Expect(fakeCFClient.GetServicePlansCallCount()).To(Equal(1))
		Expect(fakeCFClient.GetServicePlansArgsForCall(0)).To(Equal(fakeBrokerName))

		By("getting the service instances")
		Expect(fakeCFClient.GetServiceInstancesCallCount()).To(Equal(1))
		Expect(fakeCFClient.GetServiceInstancesArgsForCall(0)).To(Equal([]string{fakePlanGUID}))

		By("calling upgrade on each upgradeable instance")
		Expect(fakeCFClient.UpgradeServiceInstanceCallCount()).Should(Equal(2))
		instanceGUID1, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(0)
		instanceGUID2, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(1)
		guids := []string{instanceGUID1, instanceGUID2}
		Expect(guids).To(ConsistOf("fake-instance-guid-1", "fake-instance-guid-2"))
	})

	It("should pass the correct information to the logger", func() {
		err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, fakeLog)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeLog.InitialTotalsCallCount()).To(Equal(1))
		actualTotal, actualUpgradable := fakeLog.InitialTotalsArgsForCall(0)
		Expect(actualTotal).To(Equal(3))
		Expect(actualUpgradable).To(Equal(2))

		Expect(fakeLog.UpgradeStartingCallCount()).To(Equal(2))
		Expect(fakeLog.UpgradeStartingArgsForCall(0)).To(Equal("fake-instance-guid-1"))
		Expect(fakeLog.UpgradeStartingArgsForCall(1)).To(Equal("fake-instance-guid-2"))

		Expect(fakeLog.UpgradeSucceededCallCount()).To(Equal(2))
		Expect(fakeLog.UpgradeFailedCallCount()).To(Equal(0))
		Expect(fakeLog.FinalTotalsCallCount()).To(Equal(1))
	})

	When("no service plans are available", func() {
		It("returns error stating no plans available", func() {
			fakeCFClient.GetServicePlansReturns([]ccapi.Plan{}, nil)

			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, fakeLog)
			Expect(err).To(MatchError(fmt.Sprintf("no service plans available for broker: %s", fakeBrokerName)))
		})
	})

	When("batch size is less that number of upgradable instances", func() {
		It("upgrades all instances", func() {
			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, fakeLog)
			Expect(err).NotTo(HaveOccurred())

			By("calling upgrade on each upgradeable instance")
			Expect(fakeCFClient.UpgradeServiceInstanceCallCount()).Should(Equal(2))
			instanceGUID1, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(0)
			instanceGUID2, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(1)
			guids := []string{instanceGUID1, instanceGUID2}
			Expect(guids).To(ConsistOf("fake-instance-guid-1", "fake-instance-guid-2"))
		})
	})

	When("there are no upgradable instances", func() {
		It("should succeed and pass the correct information to the logger", func() {
			fakeCFClient.GetServiceInstancesReturns([]ccapi.ServiceInstance{}, nil)

			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, fakeLog)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLog.PrintfCallCount()).To(Equal(2))
			Expect(fakeLog.PrintfArgsForCall(1)).To(Equal(`no instances available to upgrade`))
		})
	})

	When("getting service plans fails", func() {
		It("returns the error", func() {
			fakeCFClient.GetServicePlansReturns(nil, fmt.Errorf("plan-error"))

			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 5, fakeLog)
			Expect(err).To(MatchError("plan-error"))
		})
	})

	When("getting service instances fails", func() {
		It("returns the error", func() {
			fakeCFClient.GetServiceInstancesReturns(nil, fmt.Errorf("instance-error"))

			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 5, fakeLog)
			Expect(err).To(MatchError("instance-error"))
		})
	})

	When("an instance fails to upgrade", func() {
		BeforeEach(func() {
			fakeCFClient.UpgradeServiceInstanceReturnsOnCall(0, nil)
			fakeCFClient.UpgradeServiceInstanceReturnsOnCall(1, fmt.Errorf("failed to upgrade instance"))
		})

		It("should pass the correct information to the logger", func() {
			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, fakeLog)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLog.InitialTotalsCallCount()).To(Equal(1))
			actualTotal, actualUpgradable := fakeLog.InitialTotalsArgsForCall(0)
			Expect(actualTotal).To(Equal(3))
			Expect(actualUpgradable).To(Equal(2))

			Expect(fakeLog.UpgradeStartingCallCount()).To(Equal(2))
			Expect(fakeLog.UpgradeStartingArgsForCall(0)).To(Equal("fake-instance-guid-1"))
			Expect(fakeLog.UpgradeStartingArgsForCall(1)).To(Equal("fake-instance-guid-2"))

			Expect(fakeLog.UpgradeSucceededCallCount()).To(Equal(1))
			Expect(fakeLog.UpgradeFailedCallCount()).To(Equal(1))
			Expect(fakeLog.FinalTotalsCallCount()).To(Equal(1))
		})
	})
})
