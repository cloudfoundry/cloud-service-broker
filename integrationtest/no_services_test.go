package integrationtest_test

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("No services", func() {
	It("fails to start", func() {
		brokerpak := GinkgoT().TempDir()
		broker, err := testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter))
		Expect(broker.Stop()).To(Succeed()) // clean up just in case
		Expect(err).To(MatchError(ContainSubstring("no services are defined")))
	})
})
