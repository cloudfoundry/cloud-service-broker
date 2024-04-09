package integrationtest_test

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/v2/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/testdrive"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Catalog", func() {
	const userProvidedPlan = `[{"name": "user-plan","id":"8b52a460-b246-11eb-a8f5-d349948e2480"}]`

	var brokerpak string

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("service-catalog")))

		DeferCleanup(func() {
			cleanup(brokerpak)
		})
	})

	When("a service offering has duplicate plan IDs", func() {
		It("fails to start", func() {
			_, err := testdrive.StartBroker(csb, brokerpak, database, testdrive.WithEnv(fmt.Sprintf("GSB_SERVICE_ALPHA_SERVICE_PLANS=%s", userProvidedPlan)), testdrive.WithOutputs(GinkgoWriter, GinkgoWriter))
			Expect(err).To(MatchError(ContainSubstring("duplicated value, must be unique: 8b52a460-b246-11eb-a8f5-d349948e2480: Plans[1].Id")))
		})
	})

	When("two service offerings have duplicate plan IDs", func() {
		It("fails to start", func() {
			_, err := testdrive.StartBroker(csb, brokerpak, database, testdrive.WithEnv(fmt.Sprintf("GSB_SERVICE_BETA_SERVICE_PLANS=%s", userProvidedPlan)), testdrive.WithOutputs(GinkgoWriter, GinkgoWriter))
			Expect(err).To(MatchError(ContainSubstring("duplicated value, must be unique: 8b52a460-b246-11eb-a8f5-d349948e2480: services[1].Plans[1].Id")))
		})
	})
})
