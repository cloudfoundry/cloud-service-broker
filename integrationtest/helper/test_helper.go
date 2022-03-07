package helper

import (
	"fmt"
	"os"
	"path"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/pborman/uuid"
)

func New(csb string) *TestHelper {
	tmpDir := ginkgo.GinkgoT().TempDir()

	original, err := os.Getwd()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	fmt.Fprintf(ginkgo.GinkgoWriter, "running test in: %s\n", tmpDir)
	err = os.Chdir(tmpDir)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return &TestHelper{
		Dir:          tmpDir,
		databaseFile: path.Join(tmpDir, "databaseFile.dat"),
		OriginalDir:  original,
		csb:          csb,
		port:         freePort(),
		username:     uuid.New(),
		password:     uuid.New(),
	}
}

type TestHelper struct {
	Dir          string
	OriginalDir  string
	csb          string
	databaseFile string
	port         int
	username     string
	password     string
}
