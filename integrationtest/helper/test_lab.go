package helper

import (
	"os"
	"path"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/pborman/uuid"
)

func NewTestLab(csb string) *TestLab {
	tmpDir := ginkgo.GinkgoT().TempDir()

	err := os.Chdir(tmpDir)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return &TestLab{
		csb:          csb,
		Dir:          tmpDir,
		port:         freePort(),
		username:     uuid.New(),
		password:     uuid.New(),
		DatabaseFile: path.Join(tmpDir, "DatabaseFile.dat"),
	}
}

type TestLab struct {
	Dir          string
	csb          string
	port         int
	username     string
	password     string
	DatabaseFile string
}
