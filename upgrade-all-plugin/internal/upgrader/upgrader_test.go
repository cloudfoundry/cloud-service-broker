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
		fakeCCAPI                                           *upgraderfakes.FakeCCAPI
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

		fakeCCAPI = &upgraderfakes.FakeCCAPI{}
		fakeCCAPI.GetServicePlansReturns([]ccapi.Plan{fakePlan}, nil)
		fakeCCAPI.GetServiceInstancesReturns(fakeServiceInstances, nil)

		fakeStdout = NewBuffer()
		fakeLog = log.New(fakeStdout, "", 0)
	})

	It("upgrades a service instance", func() {
		err := upgrader.Upgrade(fakeCCAPI, fakeBrokerName, 5, fakeLog)
		Expect(err).NotTo(HaveOccurred())

		By("getting the service plans")
		Expect(fakeCCAPI.GetServicePlansCallCount()).To(Equal(1))
		Expect(fakeCCAPI.GetServicePlansArgsForCall(0)).To(Equal(fakeBrokerName))

		By("getting the service instances")
		Expect(fakeCCAPI.GetServiceInstancesCallCount()).To(Equal(1))
		Expect(fakeCCAPI.GetServiceInstancesArgsForCall(0)).To(Equal([]string{fakePlanGUID}))

		By("calling upgrade on each upgradeable instance")
		Expect(fakeCCAPI.UpgradeServiceInstanceCallCount()).Should(Equal(2))
		instanceGuid1, _ := fakeCCAPI.UpgradeServiceInstanceArgsForCall(0)
		instanceGuid2, _ := fakeCCAPI.UpgradeServiceInstanceArgsForCall(1)
		guids := []string{instanceGuid1, instanceGuid2}
		Expect(guids).To(ConsistOf("fake-instance-guid-1", "fake-instance-guid-2"))
	})

	When("batch size is less that number of upgradable instances", func() {
		It("upgrades all instances", func() {
			err := upgrader.Upgrade(fakeCCAPI, fakeBrokerName, 1, fakeLog)
			Expect(err).NotTo(HaveOccurred())

			By("calling upgrade on each upgradeable instance")
			Expect(fakeCCAPI.UpgradeServiceInstanceCallCount()).Should(Equal(2))
			instanceGuid1, _ := fakeCCAPI.UpgradeServiceInstanceArgsForCall(0)
			instanceGuid2, _ := fakeCCAPI.UpgradeServiceInstanceArgsForCall(1)
			guids := []string{instanceGuid1, instanceGuid2}
			Expect(guids).To(ConsistOf("fake-instance-guid-1", "fake-instance-guid-2"))
		})
	})

	When("getting service plans fails", func() {
		It("returns the error", func() {
			fakeCCAPI.GetServicePlansReturns(nil, fmt.Errorf("plan-error"))

			err := upgrader.Upgrade(fakeCCAPI, fakeBrokerName, 5, fakeLog)
			Expect(err).To(MatchError("plan-error"))
		})
	})

	When("getting service instances fails", func() {
		It("returns the error", func() {
			fakeCCAPI.GetServiceInstancesReturns(nil, fmt.Errorf("instance-error"))

			err := upgrader.Upgrade(fakeCCAPI, fakeBrokerName, 5, fakeLog)
			Expect(err).To(MatchError("instance-error"))
		})
	})

	Describe("logging", func() {
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
				upgrader.Upgrade(fakeCCAPI, fakeBrokerName, 5, fakeLog)
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
				fakeCCAPI.UpgradeServiceInstanceReturnsOnCall(0, nil)
				fakeCCAPI.UpgradeServiceInstanceReturnsOnCall(1, fmt.Errorf("failed to upgrade instance"))
			})

			It("should output the correct logging", func() {
				upgrader.Upgrade(fakeCCAPI, fakeBrokerName, 1, fakeLog)
				Expect(fakeStdout).To(matchLogOutput)
			})
		})
	})
})
