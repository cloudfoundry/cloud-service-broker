package helper

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func (tl *TestLab) StartBrokerCommand(env ...string) *exec.Cmd {
	cmd := exec.Command(tl.csb, "serve")
	cmd.Env = append(
		os.Environ(),
		"CSB_LISTENER_HOST=localhost",
		"DB_TYPE=sqlite3",
		fmt.Sprintf("DB_PATH=%s", tl.DatabaseFile),
		fmt.Sprintf("PORT=%d", tl.port),
		fmt.Sprintf("SECURITY_USER_NAME=%s", tl.username),
		fmt.Sprintf("SECURITY_USER_PASSWORD=%s", tl.password),
	)
	cmd.Env = append(cmd.Env, env...)

	return cmd
}

func (tl *TestLab) StartBrokerSession(env ...string) *gexec.Session {
	cmd := tl.StartBrokerCommand(env...)
	session, err := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return session
}

func (tl *TestLab) StartBroker(env ...string) *gexec.Session {
	session := tl.StartBrokerSession(env...)
	waitForBrokerToStart(tl.port)
	return session
}
