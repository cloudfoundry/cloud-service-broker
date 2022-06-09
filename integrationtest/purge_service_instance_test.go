package integrationtest_test

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
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
		testHelper := helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "purge-service-instance")
		brokerSession := testHelper.StartBroker()
		DeferCleanup(func() {
			brokerSession.Terminate().Wait()
		})

		By("creating a service to keep")
		keepInstance := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
		keepBinding1GUID, _ := testHelper.CreateBinding(keepInstance, bindParams)
		keepBinding2GUID, _ := testHelper.CreateBinding(keepInstance, bindParams)

		By("creating a service without bindings to purge")
		purgeInstanceWithoutBindings := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)

		By("creating a service with bindings to purge")
		purgeInstanceWithBindings := testHelper.Provision(serviceOfferingGUID, servicePlanGUID)
		purgeBinding1GUID, _ := testHelper.CreateBinding(purgeInstanceWithBindings, bindParams)
		purgeBinding2GUID, _ := testHelper.CreateBinding(purgeInstanceWithBindings, bindParams)

		By("stopping the broker")
		brokerSession.Terminate().Wait()

		By("purging the service instances")
		purgeServiceInstance(testHelper, purgeInstanceWithBindings.GUID)
		purgeServiceInstance(testHelper, purgeInstanceWithoutBindings.GUID)

		By("checking that we purged the service instances")
		expectServiceInstanceStatus(testHelper, purgeInstanceWithoutBindings.GUID, BeFalse())
		expectServiceInstanceStatus(testHelper, purgeInstanceWithBindings.GUID, BeFalse())
		expectServiceBindingStatus(testHelper, purgeInstanceWithBindings.GUID, purgeBinding1GUID, BeFalse())
		expectServiceBindingStatus(testHelper, purgeInstanceWithBindings.GUID, purgeBinding2GUID, BeFalse())

		By("checking that the other service instance still exists")
		expectServiceInstanceStatus(testHelper, keepInstance.GUID, BeTrue())
		expectServiceBindingStatus(testHelper, keepInstance.GUID, keepBinding1GUID, BeTrue())
		expectServiceBindingStatus(testHelper, keepInstance.GUID, keepBinding2GUID, BeTrue())
	})
})

func purgeServiceInstance(testHelper *helper.TestHelper, serviceInstanceGUID string) {
	cmd := exec.Command(csb, "purge", serviceInstanceGUID)
	cmd.Env = append(
		os.Environ(),
		"DB_TYPE=sqlite3",
		fmt.Sprintf("DB_PATH=%s", testHelper.DatabaseFile),
	)
	purgeSession, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).WithOffset(1).NotTo(HaveOccurred())
	Eventually(purgeSession).WithOffset(1).Should(Exit(0))
}

func expectServiceInstanceStatus(testHelper *helper.TestHelper, guid string, match types.GomegaMatcher) {
	Expect(existsDatabaseEntry(testHelper, &models.ServiceInstanceDetails{}, "id=?", guid)).WithOffset(1).To(match, "service instance details")
	Expect(existsDatabaseEntry(testHelper, &models.ProvisionRequestDetails{}, "service_instance_id=?", guid)).WithOffset(1).To(match, "provision request details")
	Expect(existsDatabaseEntry(testHelper, &models.TerraformDeployment{}, "id=?", fmt.Sprintf("tf:%s:", guid))).WithOffset(1).To(match, "terraform deployment")
}

func expectServiceBindingStatus(testHelper *helper.TestHelper, serviceInstanceGUID, bindingGUID string, match types.GomegaMatcher) {
	Expect(existsDatabaseEntry(testHelper, &models.BindRequestDetails{}, "service_binding_id=?", bindingGUID)).WithOffset(1).To(match, "bind request details")
	Expect(existsDatabaseEntry(testHelper, &models.ServiceBindingCredentials{}, "binding_id=?", bindingGUID)).WithOffset(1).To(match, "service binding credentials")
	Expect(existsDatabaseEntry(testHelper, &models.TerraformDeployment{}, "id=?", fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, bindingGUID))).WithOffset(1).To(match, "terraform deployment")
}

func existsDatabaseEntry[T any, PtrT *T](testHelper *helper.TestHelper, model PtrT, query string, args ...any) bool {
	var count int64
	Expect(testHelper.DBConn().Model(model).Where(query, args...).Count(&count).Error).NotTo(HaveOccurred())
	return count != 0
}
