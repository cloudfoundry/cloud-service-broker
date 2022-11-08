package integrationtest_test

import (
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"

	"gorm.io/gorm"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Test Suite")
}

var (
	csb      string
	fixtures func(string) string
	database string
	dbConn   *gorm.DB
)

var _ = SynchronizedBeforeSuite(
	func() []byte {
		// -gcflags enabled "gops", but had to be removed as this doesn't compile with Go 1.19
		//path, err := Build("github.com/cloudfoundry/cloud-service-broker", `-gcflags="all=-N -l"`)
		path, err := Build("github.com/cloudfoundry/cloud-service-broker")
		Expect(err).NotTo(HaveOccurred())
		return []byte(path)
	},
	func(data []byte) {
		csb = string(data)

		cwd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		fixtures = func(name string) string {
			return filepath.Join(cwd, "fixtures", name)
		}
	},
)

var _ = SynchronizedAfterSuite(
	func() {},
	func() { CleanupBuildArtifacts() },
)

var _ = BeforeEach(func() {
	fh, err := os.CreateTemp(os.TempDir(), "csbdb")
	Expect(err).NotTo(HaveOccurred())
	defer fh.Close()

	database = fh.Name()
	DeferCleanup(func() {
		cleanup(database)
	})

	dbConn, err = gorm.Open(sqlite.Open(database), &gorm.Config{})
	Expect(err).NotTo(HaveOccurred())
})

func requestID() string {
	return uuid.New()
}

func must[A any](a A, err error) A {
	Expect(err).WithOffset(1).NotTo(HaveOccurred())
	return a
}

func cleanup(path string) {
	Expect(os.RemoveAll(path)).To(Succeed())
}
