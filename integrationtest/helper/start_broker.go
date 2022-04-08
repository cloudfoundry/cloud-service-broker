package helper

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func (h *TestHelper) StartBrokerCommand(env ...string) *exec.Cmd {
	cmd := exec.Command(h.csb, "serve")
	cmd.Env = append(
		os.Environ(),
		"CSB_LISTENER_HOST=localhost",
		"DB_TYPE=sqlite3",
		fmt.Sprintf("DB_PATH=%s", h.databaseFile),
		fmt.Sprintf("PORT=%d", h.Port),
		fmt.Sprintf("SECURITY_USER_NAME=%s", h.username),
		fmt.Sprintf("SECURITY_USER_PASSWORD=%s", h.password),
	)
	cmd.Env = append(cmd.Env, env...)

	return cmd
}

func (h *TestHelper) StartBrokerSession(env ...string) *gexec.Session {
	cmd := h.StartBrokerCommand(env...)
	session, err := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return session
}

func (h *TestHelper) StartBroker(env ...string) *gexec.Session {
	session := h.StartBrokerSession(env...)
	waitForBrokerToStart(h.Port)
	return session
}
