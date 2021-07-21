package integrationtest_test

import (
	"testing"

	. "github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brokerpak Integration Test Suite")
}

var csb string

var _ = BeforeSuite(func() {
	var err error
	csb, err = Build("github.com/cloudfoundry-incubator/cloud-service-broker")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	CleanupBuildArtifacts()
})
