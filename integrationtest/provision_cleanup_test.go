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

var _ = Describe("Provision Cleanup", func() {
	const (
		serviceOfferingGUID = "cfeda8d0-cbf3-11ee-be53-73f17d1c612b"
		servicePlanGUID     = "d8fbab66-cbf3-11ee-ab90-d7299e1fcf96"
	)

	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("provision-cleanup")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database, testdrive.WithOutputs(GinkgoWriter, GinkgoWriter)))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	// This test captures an incorrect behavior that we want to fix
	It("fails to clean up after a provision failed with empty state", func() {
		By("failing to provision")
		instance, err := broker.Provision(serviceOfferingGUID, servicePlanGUID)
		Expect(err).To(MatchError("provision failed with state: failed"))

		By("failing to clean up")
		Expect(broker.Deprovision(instance)).To(MatchError(ContainSubstring("failed to delete: workspace state not generated")))
	})

	// This test captures an incorrect behavior that we want to fix
	It("fails to clean up after a provision failed with corrupted state", func() {
		By("failing to provision")
		instance, err := broker.Provision(serviceOfferingGUID, servicePlanGUID)
		Expect(err).To(MatchError("provision failed with state: failed"))

		By("corrupting the state as if terraform had been killed")
		invalidWorkspace := must(json.Marshal(workspace.TerraformWorkspace{State: []byte(`{"foo`)})) // Base64-encoded truncated JSON
		Expect(
			dbConn.Model(&models.TerraformDeployment{}).
				Where("id = ?", fmt.Sprintf("tf:%s:", instance.GUID)).
				Update("workspace", invalidWorkspace).
				Error,
		).To(Succeed())

		By("failing to clean up")
		Expect(broker.Deprovision(instance)).To(MatchError(ContainSubstring("failed to delete: invalid workspace state unexpected end of JSON input")))
	})
})
