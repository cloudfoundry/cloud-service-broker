package integrationtest_test

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/types"
)

var _ = Describe("Purge Service Instance", func() {
	const (
		serviceOfferingGUID = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		servicePlanGUID     = "8b52a460-b246-11eb-a8f5-d349948e2480"
		bindParams          = `{"foo":"bar"}`
	)

	It("purges the correct service instance and no others", func() {
		By("creating a broker with brokerpak")
		brokerpak := must(testdrive.BuildBrokerpak(csb, fixtures("purge-service-instance")))
		broker := must(testdrive.StartBroker(csb, brokerpak, database))
		DeferCleanup(func() {
			broker.Stop()
			cleanup(brokerpak)
		})

		By("creating a service to keep")
		keepInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
		keepBinding1 := must(broker.CreateBinding(keepInstance, testdrive.WithBindingParams(bindParams)))
		keepBinding2 := must(broker.CreateBinding(keepInstance, testdrive.WithBindingParams(bindParams)))

		By("creating a service without bindings to purge")
		purgeInstanceWithoutBindings := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))

		By("creating a service with bindings to purge")
		purgeInstanceWithBindings := must(broker.Provision(serviceOfferingGUID, servicePlanGUID))
		purgeBinding1 := must(broker.CreateBinding(purgeInstanceWithBindings, testdrive.WithBindingParams(bindParams)))
		purgeBinding2 := must(broker.CreateBinding(purgeInstanceWithBindings, testdrive.WithBindingParams(bindParams)))

		By("stopping the broker")
		broker.Stop()

		By("purging the service instances")
		purgeServiceInstance(database, purgeInstanceWithBindings.GUID)
		purgeServiceInstance(database, purgeInstanceWithoutBindings.GUID)

		By("checking that we purged the service instances")
		expectServiceInstanceStatus(purgeInstanceWithoutBindings.GUID, BeFalse())
		expectServiceInstanceStatus(purgeInstanceWithBindings.GUID, BeFalse())
		expectServiceBindingStatus(purgeInstanceWithBindings.GUID, purgeBinding1.GUID, BeFalse())
		expectServiceBindingStatus(purgeInstanceWithBindings.GUID, purgeBinding2.GUID, BeFalse())

		By("checking that the other service instance still exists")
		expectServiceInstanceStatus(keepInstance.GUID, BeTrue())
		expectServiceBindingStatus(keepInstance.GUID, keepBinding1.GUID, BeTrue())
		expectServiceBindingStatus(keepInstance.GUID, keepBinding2.GUID, BeTrue())
	})
})

func purgeServiceInstance(database, serviceInstanceGUID string) {
	cmd := exec.Command(csb, "purge", serviceInstanceGUID)
	cmd.Env = append(
		os.Environ(),
		"DB_TYPE=sqlite3",
		fmt.Sprintf("DB_PATH=%s", database),
	)
	purgeSession, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).WithOffset(1).NotTo(HaveOccurred())
	Eventually(purgeSession).WithTimeout(time.Minute).WithOffset(1).Should(Exit(0))
}

func expectServiceInstanceStatus(guid string, match types.GomegaMatcher) {
	Expect(existsDatabaseEntry(&models.ServiceInstanceDetails{}, "id=?", guid)).WithOffset(1).To(match, "service instance details")
	Expect(existsDatabaseEntry(&models.ProvisionRequestDetails{}, "service_instance_id=?", guid)).WithOffset(1).To(match, "provision request details")
	Expect(existsDatabaseEntry(&models.TerraformDeployment{}, "id=?", fmt.Sprintf("tf:%s:", guid))).WithOffset(1).To(match, "terraform deployment")
}

func expectServiceBindingStatus(serviceInstanceGUID, bindingGUID string, match types.GomegaMatcher) {
	Expect(existsDatabaseEntry(&models.BindRequestDetails{}, "service_binding_id=?", bindingGUID)).WithOffset(1).To(match, "bind request details")
	Expect(existsDatabaseEntry(&models.ServiceBindingCredentials{}, "binding_id=?", bindingGUID)).WithOffset(1).To(match, "service binding credentials")
	Expect(existsDatabaseEntry(&models.TerraformDeployment{}, "id=?", fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, bindingGUID))).WithOffset(1).To(match, "terraform deployment")
}

func existsDatabaseEntry[T any, PtrT *T](model PtrT, query string, args ...any) bool {
	var count int64
	Expect(dbConn.Model(model).Where(query, args...).Count(&count).Error).NotTo(HaveOccurred())
	return count != 0
}
