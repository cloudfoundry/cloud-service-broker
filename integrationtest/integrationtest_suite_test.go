package integrationtest_test

import (
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"

	"gorm.io/gorm"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
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
		// path, err := Build("github.com/cloudfoundry/cloud-service-broker", `-gcflags="all=-N -l"`)
		path := must(Build("github.com/cloudfoundry/cloud-service-broker/v3"))
		return []byte(path)
	},
	func(data []byte) {
		csb = string(data)

		cwd := must(os.Getwd())
		fixtures = func(name string) string {
			return filepath.Join(cwd, "fixtures", name)
		}
	},
)

var _ = SynchronizedAfterSuite(
	func() {
	},
	func() {
		CleanupBuildArtifacts()
		files, err := filepath.Glob("/tmp/brokerpak*")
		Expect(err).ToNot(HaveOccurred())
		for _, f := range files {
			if err := os.RemoveAll(f); err != nil {
				Expect(err).ToNot(HaveOccurred())
			}
		}
	},
)

var _ = BeforeEach(func() {
	fh := must(os.CreateTemp(os.TempDir(), "csbdb"))
	defer fh.Close()

	database = fh.Name()
	DeferCleanup(func() {
		cleanup(database)
	})

	dbConn = must(gorm.Open(sqlite.Open(database), &gorm.Config{}))
})

func requestID() string {
	return uuid.NewString()
}

func must[A any](a A, err error) A {
	Expect(err).WithOffset(1).NotTo(HaveOccurred())
	return a
}

func cleanup(path string) {
	Expect(os.RemoveAll(path)).To(Succeed())
}
