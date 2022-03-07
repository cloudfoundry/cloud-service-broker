package helper

import (
	"fmt"
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func (h *TestHelper) Restore() {
	fmt.Fprintf(ginkgo.GinkgoWriter, "switching back to: %s\n", h.OriginalDir)
	err := os.Chdir(h.OriginalDir)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}
