package helper

import (
	"os/exec"
	"path"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func (h *TestHelper) BuildBrokerpakCommand(paths ...string) *exec.Cmd {
	return exec.Command(h.csb, "pak", "build", path.Join(paths...))
}

func (h *TestHelper) BuildBrokerpak(paths ...string) {
	session, err := gexec.Start(h.BuildBrokerpakCommand(paths...), ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
}
