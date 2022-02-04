package helper

import (
	"os"

	"github.com/onsi/gomega"
)

func OriginalDir() Original {
	dir, err := os.Getwd()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return Original(dir)
}

type Original string

func (o Original) Return() {
	err := os.Chdir(string(o))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}
