package integrationtest_test

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Purge Service Binding", func() {
	const (
		serviceOfferingGUID = "2f36d5c6-ccc3-11ee-a3be-cb7c74dcfe9a"
		servicePlanGUID     = "21a3e6c4-ccc3-11ee-a9dd-d74726b3c0d2"
		bindParams          = `{"foo":"bar"}`
	)

	It("purges the correct service binding and no others", func() {
		By("creating a broker with brokerpak")
		brokerpak := must(packer.BuildBrokerpak(csb, fixtures("purge-service-binding")))
		broker := must(testdrive.StartBroker(csb, brokerpak, database))
		DeferCleanup(func() {
			broker.Stop()
			cleanup(brokerpak)
		})

		By("creating a service with bindings to purge")
		instance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
		keepBinding1 := must(broker.CreateBinding(instance, testdrive.WithBindingParams(bindParams)))
		purgeBinding := must(broker.CreateBinding(instance, testdrive.WithBindingParams(bindParams)))
		keepBinding2 := must(broker.CreateBinding(instance, testdrive.WithBindingParams(bindParams)))

		By("stopping the broker")
		broker.Stop()

		By("purging the binding")
		purgeServiceBinding(database, instance.GUID, purgeBinding.GUID)

		By("checking that we purged the service binding")
		expectServiceBindingStatus(instance.GUID, purgeBinding.GUID, BeFalse())

		By("checking that the other service bindings still exists")
		expectServiceBindingStatus(instance.GUID, keepBinding1.GUID, BeTrue())
		expectServiceBindingStatus(instance.GUID, keepBinding2.GUID, BeTrue())
	})
})

func purgeServiceBinding(database, serviceInstanceGUID, serviceBindingGUID string) {
	cmd := exec.Command(csb, "purge-binding", serviceInstanceGUID, serviceBindingGUID)
	cmd.Env = append(
		os.Environ(),
		"DB_TYPE=sqlite3",
		fmt.Sprintf("DB_PATH=%s", database),
	)
	purgeSession, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).WithOffset(1).NotTo(HaveOccurred())
	Eventually(purgeSession).WithTimeout(time.Minute).WithOffset(1).Should(Exit(0))
}
