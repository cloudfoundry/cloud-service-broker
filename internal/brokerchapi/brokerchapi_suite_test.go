package brokerchapi_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestBrokerCHAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Broker CredHub API Suite")
}

var fakeCertsPath string

var _ = SynchronizedBeforeSuite(
	func() []byte {
		path := GinkgoT().TempDir()

		cmd := exec.Command("bash", filepath.Join(must(os.Getwd()), "generateFakeCerts.sh"))
		cmd.Dir = path
		session := must(gexec.Start(cmd, GinkgoWriter, GinkgoWriter))
		Eventually(session).WithTimeout(time.Minute).WithPolling(time.Second).Should(gexec.Exit(0))
		return []byte(path)
	},
	func(input []byte) {
		fakeCertsPath = string(input)
	},
)
