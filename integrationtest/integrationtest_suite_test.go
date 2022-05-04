package integrationtest_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	"github.com/pivotal-cf/brokerapi/v8/domain"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Test Suite")
}

var csb string

var lastOperationPollingFrequency = time.Second * 1

var _ = SynchronizedBeforeSuite(
	func() []byte {
		path, err := Build("github.com/cloudfoundry/cloud-service-broker", `-gcflags="all=-N -l"`)
		Expect(err).NotTo(HaveOccurred())
		return []byte(path)
	},
	func(data []byte) {
		csb = string(data)
	},
)

var _ = SynchronizedAfterSuite(
	func() {},
	func() { CleanupBuildArtifacts() },
)

func requestID() string {
	return uuid.New()
}

func pollLastOperation(testHelper *helper.TestHelper, serviceInstanceGUID string) func() domain.LastOperationState {
	return func() domain.LastOperationState {
		lastOperationResponse := testHelper.Client().LastOperation(serviceInstanceGUID, requestID())
		Expect(lastOperationResponse.Error).NotTo(HaveOccurred())
		Expect(lastOperationResponse.StatusCode).To(Or(Equal(http.StatusOK), Equal(http.StatusGone)))
		var receiver domain.LastOperation
		err := json.Unmarshal(lastOperationResponse.ResponseBody, &receiver)
		Expect(err).NotTo(HaveOccurred())
		return receiver.State
	}
}
