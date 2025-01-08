package integrationtest_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/testdrive"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v12/domain"
)

//go:embed "fixtures/import-state-data/terraform.tfstate"
var stateToImport []byte

var _ = Describe("Import State", func() {
	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	BeforeEach(func() {
		brokerpak = must(packer.BuildBrokerpak(csb, fixtures("import-state")))
		broker = must(testdrive.StartBroker(csb, brokerpak, database))

		DeferCleanup(func() {
			Expect(broker.Terminate()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	It("can create a vacant service instance and import a terraform state", func() {
		const (
			serviceOfferingGUID = "5b4f6244-f7ee-11ee-b5b3-3389c8712346"
			servicePlanGUID     = "5b50951a-f7ee-11ee-b564-6b989de50807"
			importedValue       = "831e0be4-7ff3-26ac-751c-80951adb3fe7" // matches what's in the test fixture
		)

		By("creating a 'vacant' service instance")
		instance, err := broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(`{"vacant":true}`))
		Expect(err).NotTo(HaveOccurred())

		By("checking that the state is empty")
		var d models.TerraformDeployment
		terraformDeploymentID := fmt.Sprintf("tf:%s:", instance.GUID)
		Expect(dbConn.Where("id = ?", terraformDeploymentID).First(&d).Error).To(Succeed())
		var w workspace.TerraformWorkspace
		Expect(json.Unmarshal(d.Workspace, &w)).To(Succeed())
		Expect(w.State).To(MatchJSON(`{"version":4}`))

		By("checking that the `vacant` parameter was not stored")
		var i models.ProvisionRequestDetails
		Expect(dbConn.Where("service_instance_id = ?", instance.GUID).First(&i).Error).To(Succeed())
		Expect(i.RequestDetails).To(MatchJSON(`{}`))

		By("importing state into the vacant service instance")
		req := must(http.NewRequest(http.MethodPatch, fmt.Sprintf("http://localhost:%d/import_state/%s", broker.Port, instance.GUID), bytes.NewReader(stateToImport)))
		req.SetBasicAuth(broker.Username, broker.Password)
		importResponse := must(broker.Client.Do(req))
		Expect(importResponse).To(HaveHTTPStatus(http.StatusOK))

		By("checking that the state was imported into the database")
		Expect(dbConn.Where("id = ?", terraformDeploymentID).First(&d).Error).To(Succeed())
		Expect(json.Unmarshal(d.Workspace, &w)).To(Succeed())
		Expect(w.State).To(MatchJSON(stateToImport))

		By("performing a no-op update to trigger an Apply")
		updateResponse := broker.Client.Update(instance.GUID, instance.ServiceOfferingGUID, servicePlanGUID, uuid.NewString(), nil, domain.PreviousValues{}, nil)
		Expect(updateResponse.Error).NotTo(HaveOccurred())
		Expect(updateResponse.StatusCode).To(Equal(http.StatusAccepted))
		state, err := broker.LastOperationFinalValue(instance.GUID)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.State).To(BeEquivalentTo("succeeded"))

		By("checking that data from the imported state made it into the Last Operation output")
		Expect(state.Description).To(Equal(fmt.Sprintf("update succeeded: created random GUID: %s", importedValue)))
	})
})
