package helper

import (
	"os"

	"github.com/onsi/gomega"
)

func (h *TestHelper) Restore() {
	err := os.Chdir(h.OriginalDir)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}
