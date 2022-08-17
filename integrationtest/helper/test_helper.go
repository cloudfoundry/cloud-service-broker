package helper

import (
	"fmt"
	"os"
	"path"

	"github.com/cloudfoundry/cloud-service-broker/utils/freeport"
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

	ginkgo.DeferCleanup(func() {
		fmt.Fprintf(ginkgo.GinkgoWriter, "switching back to: %s\n", original)
		err := os.Chdir(original)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	return &TestHelper{
		Dir:          tmpDir,
		DatabaseFile: path.Join(tmpDir, "databaseFile.dat"),
		OriginalDir:  original,
		csb:          csb,
		Port:         freeport.Must(),
		username:     uuid.New(),
		password:     uuid.New(),
	}
}

type TestHelper struct {
	Dir          string
	OriginalDir  string
	Port         int
	DatabaseFile string
	csb          string
	username     string
	password     string
}
