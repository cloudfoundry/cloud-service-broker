package scheduler_test

import (
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/ccapi"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/scheduler"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("ScheduleUpgrades", func() {
	const testPlanGUID1 = "test-plan-guid-1"
	const testPlanGUID2 = "test-plan-guid-2"

	var (
		upgradeQueue chan scheduler.UpgradeTask
		instances    []ccapi.ServiceInstance
		planVersions map[string]string
	)

	BeforeEach(func() {
		upgradeQueue = make(chan scheduler.UpgradeTask)
		instances = []ccapi.ServiceInstance{
			{
				GUID:             "test-guid-1",
				UpgradeAvailable: true,
				Relationships: struct {
					ServicePlan struct {
						Data struct {
							GUID string `json:"guid"`
						} `json:"data"`
					} `json:"service_plan"`
				}{
					ServicePlan: struct {
						Data struct {
							GUID string `json:"guid"`
						} `json:"data"`
					}{
						Data: struct {
							GUID string `json:"guid"`
						}{
							GUID: testPlanGUID1,
						},
					},
				},
			},
			{
				GUID:             "test-guid-2",
				UpgradeAvailable: true,
				Relationships: struct {
					ServicePlan struct {
						Data struct {
							GUID string `json:"guid"`
						} `json:"data"`
					} `json:"service_plan"`
				}{
					ServicePlan: struct {
						Data struct {
							GUID string `json:"guid"`
						} `json:"data"`
					}{
						Data: struct {
							GUID string `json:"guid"`
						}{
							GUID: testPlanGUID2,
						},
					},
				},
			},
			{
				GUID:             "test-guid-3",
				UpgradeAvailable: true,
				Relationships: struct {
					ServicePlan struct {
						Data struct {
							GUID string `json:"guid"`
						} `json:"data"`
					} `json:"service_plan"`
				}{
					ServicePlan: struct {
						Data struct {
							GUID string `json:"guid"`
						} `json:"data"`
					}{
						Data: struct {
							GUID string `json:"guid"`
						}{
							GUID: testPlanGUID1,
						},
					},
				},
			},
			{
				GUID:             "test-guid-4",
				UpgradeAvailable: false,
				Relationships: struct {
					ServicePlan struct {
						Data struct {
							GUID string `json:"guid"`
						} `json:"data"`
					} `json:"service_plan"`
				}{
					ServicePlan: struct {
						Data struct {
							GUID string `json:"guid"`
						} `json:"data"`
					}{
						Data: struct {
							GUID string `json:"guid"`
						}{
							GUID: testPlanGUID1,
						},
					},
				},
			},
		}
		planVersions = map[string]string{
			testPlanGUID1: "1.0.0",
			testPlanGUID2: "2.0.0",
		}
	})

	It("assigns the correct plan version to an upgrade task", func() {
		scheduler.ScheduleUpgrades(upgradeQueue, instances, planVersions)

	})
	It("only adds instances where upgrade is available", func() {

	})
	It("closes the given channel when all items are read from", func() {})

})
