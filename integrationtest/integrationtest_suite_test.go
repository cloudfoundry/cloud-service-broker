package integrationtest_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Test Suite")
}

var csb string

var _ = SynchronizedBeforeSuite(
	func() []byte {
		// -gcflags enabled "gops", but had to be removed as this doesn't compile with Go 1.19
		//path, err := Build("github.com/cloudfoundry/cloud-service-broker", `-gcflags="all=-N -l"`)
		path, err := Build("github.com/cloudfoundry/cloud-service-broker")
		Expect(err).NotTo(HaveOccurred())
		return []byte(path)
	},
	func(data []byte) {
		csb = string(data)
	},
)

var _ = SynchronizedAfterSuite(
	func() {},
	func() { CleanupBuildArtifacts() },
)

func requestID() string {
	return uuid.New()
}
