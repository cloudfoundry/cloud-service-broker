package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudfoundry-incubator/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Plan-less", func() {
	var (
		originalDir helper.Original
		testLab     *helper.TestLab
		session     *Session
	)

	BeforeEach(func() {
		originalDir = helper.OriginalDir()
		testLab = helper.NewTestLab(csb)
	})

	AfterEach(func() {
		session.Terminate()
		originalDir.Return()
	})

	It("creates a default plan", func() {
		type catalog struct {
			Services []domain.Service `json:"services"`
		}

		var planID string

		testLab.BuildBrokerpak(string(originalDir), "fixtures", "brokerpak-without-a-plan")
		session = testLab.StartBroker()

		By("checking the catalog response", func() {
			response := testLab.Client().Catalog(requestID())
			Expect(response.Error).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			var receiver catalog
			Expect(json.Unmarshal(response.ResponseBody, &receiver)).NotTo(HaveOccurred())
			Expect(receiver.Services).To(HaveLen(1))
			Expect(receiver.Services[0].Plans).To(HaveLen(1))
			Expect(receiver.Services[0].Plans[0].Name).To(Equal("default"))
			planID = receiver.Services[0].Plans[0].ID
			Expect(planID).To(HaveLen(36))
		})

		By("checking that the UUID does not change", func() {
			session.Terminate()
			session = testLab.StartBroker()

			response := testLab.Client().Catalog(requestID())
			Expect(response.Error).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			var receiver catalog
			Expect(json.Unmarshal(response.ResponseBody, &receiver)).NotTo(HaveOccurred())
			Expect(receiver.Services[0].Plans[0].ID).To(Equal(planID))
		})

		By("checking that the default plan evaporates if there is a user-defined plan", func() {
			const userProvidedPlan = `[{"name": "user-plan","id":"8b52a460-b246-11eb-a8f5-d349948e2480"}]`

			session.Terminate()
			session = testLab.StartBroker(fmt.Sprintf("GSB_SERVICE_ALPHA_SERVICE_PLANS=%s", userProvidedPlan))

			response := testLab.Client().Catalog(requestID())
			Expect(response.Error).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			var receiver catalog
			Expect(json.Unmarshal(response.ResponseBody, &receiver)).NotTo(HaveOccurred())
			Expect(receiver.Services).To(HaveLen(1))
			Expect(receiver.Services[0].Plans).To(HaveLen(1))
			Expect(receiver.Services[0].Plans[0].Name).To(Equal("user-plan"))
			Expect(receiver.Services[0].Plans[0].ID).To(Equal("8b52a460-b246-11eb-a8f5-d349948e2480"))
		})
	})
})
