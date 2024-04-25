package integrationtest_test

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/gomega/gbytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("TF Dump", func() {
	It("does not run database migration", func() {
		cmd := exec.Command(csb, "tf", "dump", "tf:fake-id:")
		cmd.Env = append(
			os.Environ(),
			"DB_TYPE=sqlite3",
			fmt.Sprintf("DB_PATH=%s", database),
		)
		session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).WithTimeout(time.Minute).Should(Exit(2))
		Expect(session.Err).To(Say("panic: no such table: password_metadata"))
	})
})
