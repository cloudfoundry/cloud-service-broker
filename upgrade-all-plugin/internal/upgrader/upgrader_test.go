package upgrader_test

import (
	"fmt"
	"log"

	. "github.com/onsi/gomega/gbytes"

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
		fakeCFClient                                        *upgraderfakes.FakeCFClient
		fakePlan                                            ccapi.Plan
		fakeInstance1, fakeInstance2, fakeInstanceNoUpgrade ccapi.ServiceInstance
		fakeServiceInstances                                []ccapi.ServiceInstance
		fakeStdout                                          *Buffer
		fakeLog                                             *log.Logger
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

		fakeStdout = NewBuffer()
		fakeLog = log.New(fakeStdout, "", 0)
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
		It("should return with no error", func() {
			fakeCFClient.GetServiceInstancesReturns([]ccapi.ServiceInstance{}, nil)
			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 5, fakeLog)
			Expect(err).NotTo(HaveOccurred())
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

	Describe("logging", func() {
		When("no instances available to upgrade", func() {
			matchLogOutput := SatisfyAll(
				Say(`no instances available to upgrade\n`),
			)

			It("should output the correct logging", func() {
				fakeCFClient.GetServiceInstancesReturns([]ccapi.ServiceInstance{}, nil)

				upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, fakeLog)
				Expect(fakeStdout).To(matchLogOutput)
			})
		})
		When("no failed upgrades", func() {
			matchLogOutput := SatisfyAll(
				Say(fmt.Sprintf(`Discovering service instances for broker: %s`, fakeBrokerName)),
				Say(`---\n`),
				Say(`Total instances: 3\n`),
				Say(`Upgradable instances: 2\n`),
				Say(`---\n`),
				Say(`Starting upgrade...`),
				Say(`---\n`),
				Say(`Finished upgrade:`),
				Say(`Total instances upgraded: 2`),
			)
			It("should output the correct logging", func() {
				upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, fakeLog)
				Expect(fakeStdout).To(matchLogOutput)
			})
		})
		When("an instance fails to upgrade", func() {
			matchLogOutput := SatisfyAll(
				Say(fmt.Sprintf(`Discovering service instances for broker: %s`, fakeBrokerName)),
				Say(`---\n`),
				Say(`Total instances: 3\n`),
				Say(`Upgradable instances: 2\n`),
				Say(`---\n`),
				Say(`Starting upgrade...`),
				Say(`---\n`),
				Say(`Finished upgrade:`),
				Say(`Total instances upgraded: 1`),
				Say(`Failed to upgrade instances:`),
				Say(`GUID	Error`),
				Say(`fake-instance-guid-2	failed to upgrade instance`),
			)
			BeforeEach(func() {
				fakeCFClient.UpgradeServiceInstanceReturnsOnCall(0, nil)
				fakeCFClient.UpgradeServiceInstanceReturnsOnCall(1, fmt.Errorf("failed to upgrade instance"))
			})

			It("should output the correct logging", func() {
				upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, fakeLog)
				Expect(fakeStdout).To(matchLogOutput)
			})
		})
	})
})
