package integrationtest_test

import (
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Catalog", func() {
	const userProvidedPlan = `[{"name": "user-plan","id":"8b52a460-b246-11eb-a8f5-d349948e2480"}]`

	var (
		testHelper *helper.TestHelper
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "brokerpak-for-catalog-test")
	})

	AfterEach(func() {
		testHelper.Restore()
	})

	When("a service offering has duplicate plan IDs", func() {
		It("fails to start", func() {
			cmd := testHelper.StartBrokerCommand(fmt.Sprintf("GSB_SERVICE_ALPHA_SERVICE_PLANS=%s", userProvidedPlan))
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(time.Minute)

			Expect(session.ExitCode()).NotTo(BeZero())
			Expect(session.Err).To(Say("duplicated value, must be unique: 8b52a460-b246-11eb-a8f5-d349948e2480: Plans\\[1\\].Id\n"))
		})
	})

	When("two service offerings have duplicate plan IDs", func() {
		It("fails to start", func() {
			cmd := testHelper.StartBrokerCommand(fmt.Sprintf("GSB_SERVICE_BETA_SERVICE_PLANS=%s", userProvidedPlan))
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(time.Minute)

			Expect(session.ExitCode()).NotTo(BeZero())
			Expect(session.Err).To(Say("duplicated value, must be unique: 8b52a460-b246-11eb-a8f5-d349948e2480: services\\[1\\].Plans\\[1\\].Id\n"))
		})
	})
})
