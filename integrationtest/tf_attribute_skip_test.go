package integrationtest_test

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("The tf_attribute_skip property", func() {
	const (
		serviceOfferingGUID = "75384ad6-48ae-11ed-a6b1-53f54b82d2aa"
		defaultPlanGUID     = "8185cfb6-48ae-11ed-8152-7bc5a2d3a884"
	)

	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("tf_attribute_skip")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	It("fails when skip is false", func() {
		s := must(broker.Provision(serviceOfferingGUID, defaultPlanGUID))
		Expect(broker.UpdateService(s)).To(MatchError(ContainSubstring(fmt.Sprintf(`error retrieving expected parameters for \"%s\": cannot find required import values for fields: does.not.exist`, s.GUID))))
	})

	It("can skip based on a stored request parameter", func() {
		s := must(broker.Provision(serviceOfferingGUID, defaultPlanGUID, testdrive.WithProvisionParams(`{"skip":true}`)))
		Expect(broker.UpdateService(s)).To(Succeed())
	})

	It("can skip based on a new request parameter", func() {
		s := must(broker.Provision(serviceOfferingGUID, defaultPlanGUID))
		Expect(broker.UpdateService(s, testdrive.WithUpdateParams(`{"skip":true}`))).To(Succeed())
	})

	It("can skip based on a plan parameter", func() {
		const skipPlanGUID = "56591d42-48af-11ed-bda0-0327763028ca"
		s := must(broker.Provision(serviceOfferingGUID, skipPlanGUID))
		Expect(broker.UpdateService(s)).To(Succeed())
	})
})
