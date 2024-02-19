package integrationtest_test

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Binding Cleanup", func() {
	const (
		badServiceOfferingGUID  = "81b4ebaa-cc08-11ee-bd34-0f8547e428e9"
		badServicePlanGUID      = "9ff671e2-cc08-11ee-bb95-3facf049ac9d"
		goodServiceOfferingGUID = "7779a92a-cc0b-11ee-85c4-4b4aa590c58a"
		goodServicePlanGUID     = "911ce91e-cc0b-11ee-a5e8-33dbc3f841a1"
	)

	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("binding-cleanup")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	It("does not need to clean up after a binding failed cleanly", func() {
		By("provisioning successfully")
		instance := must(broker.Provision(badServiceOfferingGUID, badServicePlanGUID))

		By("failing to bind")
		binding, err := broker.CreateBinding(instance)
		Expect(err).To(MatchError(ContainSubstring("error performing bind: error waiting for result: bind failed: Error: Missing required argument")))

		By("seeing an HTTP 410 Gone error")
		Expect(broker.DeleteBinding(instance, binding.GUID)).To(MatchError(ContainSubstring("unexpected status code 410")))
	})

	// This test captures an incorrect behavior that we want to fix
	It("fails to clean up after a corrupted binding", func() {
		By("provisioning successfully")
		instance := must(broker.Provision(goodServiceOfferingGUID, goodServicePlanGUID))

		By("binding successfully")
		binding := must(broker.CreateBinding(instance))

		By("corrupting the state as if terraform had been killed")
		invalidWorkspace := must(json.Marshal(workspace.TerraformWorkspace{State: []byte(`{"foo`)})) // Base64-encoded truncated JSON
		Expect(
			dbConn.Model(&models.TerraformDeployment{}).
				Where("id = ?", fmt.Sprintf("tf:%s:%s", instance.GUID, binding.GUID)).
				Update("workspace", invalidWorkspace).
				Error,
		).To(Succeed())

		By("failing to clean up the binding")
		Expect(broker.DeleteBinding(instance, binding.GUID)).To(MatchError(ContainSubstring("failed to unbind: invalid workspace state unexpected end of JSON input")))
	})
})
