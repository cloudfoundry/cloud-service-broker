package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("The tf_attribute_skip property", func() {
	const (
		serviceOfferingGUID = "75384ad6-48ae-11ed-a6b1-53f54b82d2aa"
		defaultPlanGUID     = "8185cfb6-48ae-11ed-8152-7bc5a2d3a884"
	)

	var (
		testHelper *helper.TestHelper
		session    *Session
	)

	BeforeEach(func() {
		testHelper = helper.New(csb)
		testHelper.BuildBrokerpak(testHelper.OriginalDir, "fixtures", "tf_attribute_skip")
		session = testHelper.StartBroker()
	})

	AfterEach(func() {
		session.Terminate().Wait()
	})

	It("fails when skip is false", func() {
		s := testHelper.Provision(serviceOfferingGUID, defaultPlanGUID)
		updateResponse := testHelper.Client().Update(s.GUID, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.New(), nil, domain.PreviousValues{}, nil)
		Expect(updateResponse.Error).NotTo(HaveOccurred())
		Expect(updateResponse.StatusCode).To(Equal(http.StatusInternalServerError))
		var receiver struct {
			Description string `json:"description"`
		}
		Expect(json.Unmarshal(updateResponse.ResponseBody, &receiver)).To(Succeed())
		Expect(receiver.Description).To(Equal(fmt.Sprintf(`error retrieving expected parameters for "%s": cannot find required import values for fields: does.not.exist`, s.GUID)))
	})

	It("can skip based on a stored request parameter", func() {
		s := testHelper.Provision(serviceOfferingGUID, defaultPlanGUID, `{"skip":true}`)
		testHelper.UpdateService(s)
	})

	It("can skip based on a new request parameter", func() {
		s := testHelper.Provision(serviceOfferingGUID, defaultPlanGUID, nil)
		testHelper.UpdateService(s, `{"skip":true}`)
	})

	It("can skip based on a plan parameter", func() {
		const skipPlanGUID = "56591d42-48af-11ed-bda0-0327763028ca"
		s := testHelper.Provision(serviceOfferingGUID, skipPlanGUID, nil)
		testHelper.UpdateService(s)
	})
})
