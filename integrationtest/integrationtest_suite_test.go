package integrationtest_test

import (
	"fmt"
	"net"
	"net/http"
	"testing"

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

var _ = BeforeSuite(func() {
	var err error
	csb, err = Build("github.com/cloudfoundry-incubator/cloud-service-broker")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	CleanupBuildArtifacts()
})

func freePort() int {
	listener, err := net.Listen("tcp", "localhost:0")
	Expect(err).NotTo(HaveOccurred())
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func checkAlive(port int) bool {
	response, err := http.Head(fmt.Sprintf("http://localhost:%d", port))
	return err == nil && response.StatusCode == http.StatusOK
}

func requestID() string {
	return uuid.New()
}
