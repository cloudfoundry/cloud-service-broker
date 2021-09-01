package integrationtest_test

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
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
		path, err := Build("github.com/cloudfoundry-incubator/cloud-service-broker")
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

func freePort() int {
	listener, err := net.Listen("tcp", "localhost:0")
	Expect(err).NotTo(HaveOccurred())
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func waitForBrokerToStart(port int) {
	ping := func() (*http.Response, error) {
		return http.Head(fmt.Sprintf("http://localhost:%d", port))
	}

	Eventually(ping, 30*time.Second).Should(HaveHTTPStatus(http.StatusOK))
}

func requestID() string {
	return uuid.New()
}
