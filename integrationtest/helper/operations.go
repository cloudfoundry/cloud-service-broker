package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

const (
	pollingInterval = time.Second
	timeout         = 2 * time.Minute
)

type ServiceInstance struct {
	GUID                string
	ServicePlanGUID     string
	ServiceOfferingGUID string
}

func (h *TestHelper) Provision(serviceOfferingGUID, servicePlanGUID string, params ...any) ServiceInstance {
	const offset = 1
	serviceInstanceGUID := uuid.New()
	provisionResponse := h.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), toJSONRawMessage(params, 2))
	gomega.Expect(provisionResponse.Error).WithOffset(offset).NotTo(gomega.HaveOccurred())
	gomega.Expect(provisionResponse.StatusCode).WithOffset(offset).To(gomega.Equal(http.StatusAccepted), string(provisionResponse.ResponseBody))
	gomega.Expect(h.LastOperationFinalState(serviceInstanceGUID, 1)).WithOffset(offset).To(gomega.Equal(domain.Succeeded))

	return ServiceInstance{
		GUID:                serviceInstanceGUID,
		ServicePlanGUID:     servicePlanGUID,
		ServiceOfferingGUID: serviceOfferingGUID,
	}
}

func (h *TestHelper) UpdateService(s ServiceInstance, extras ...any) {
	const offset = 1
	params := json.RawMessage(`{}`)
	var previous domain.PreviousValues
	switch len(extras) {
	case 2:
		previous = extras[1].(domain.PreviousValues)
		fallthrough
	case 1:
		params = toJSONRawMessage(extras[0:1], 2)
	case 0:
	default:
		ginkgo.Fail("too many extras")
	}

	updateResponse := h.Client().Update(s.GUID, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.New(), params, previous, nil)
	gomega.Expect(updateResponse.Error).WithOffset(offset).NotTo(gomega.HaveOccurred())
	gomega.Expect(updateResponse.StatusCode).WithOffset(offset).To(gomega.Equal(http.StatusAccepted), string(updateResponse.ResponseBody))
	gomega.Expect(h.LastOperationFinalState(s.GUID, 1)).WithOffset(offset).To(gomega.Equal(domain.Succeeded))
}

func (h *TestHelper) UpgradeService(s ServiceInstance, previousValues domain.PreviousValues, newMaintenanceInfo domain.MaintenanceInfo, params ...any) {
	const offset = 1
	updateResponse := h.Client().Update(s.GUID, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.New(), toJSONRawMessage(params, 2), previousValues, &newMaintenanceInfo)
	gomega.Expect(updateResponse.Error).WithOffset(offset).NotTo(gomega.HaveOccurred())
	gomega.Expect(updateResponse.StatusCode).WithOffset(offset).To(gomega.Equal(http.StatusAccepted), string(updateResponse.ResponseBody))
	gomega.Expect(h.LastOperationFinalState(s.GUID, 1)).WithOffset(offset).To(gomega.Equal(domain.Succeeded))
}

func (h *TestHelper) Deprovision(s ServiceInstance) {
	const offset = 1
	deprovisionResponse := h.Client().Deprovision(s.GUID, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.New())
	gomega.Expect(deprovisionResponse.Error).WithOffset(offset).NotTo(gomega.HaveOccurred())
	gomega.Expect(deprovisionResponse.StatusCode).WithOffset(offset).To(gomega.Equal(http.StatusAccepted), string(deprovisionResponse.ResponseBody))
	gomega.Expect(h.LastOperationFinalState(s.GUID, 1)).WithOffset(offset).To(gomega.Equal(domain.Succeeded))
}

func (h *TestHelper) CreateBinding(s ServiceInstance, params ...any) (string, string) {
	const offset = 1
	serviceBindingGUID := uuid.New()
	bindResponse := h.Client().Bind(s.GUID, serviceBindingGUID, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.New(), toJSONRawMessage(params, 2))
	gomega.Expect(bindResponse.Error).WithOffset(offset).NotTo(gomega.HaveOccurred())
	gomega.Expect(bindResponse.StatusCode).WithOffset(offset).To(gomega.Equal(http.StatusCreated))
	return serviceBindingGUID, string(bindResponse.ResponseBody)
}

func (h *TestHelper) DeleteBinding(s ServiceInstance, serviceBindingGUID string) {
	const offset = 1
	unbindResponse := h.Client().Unbind(s.GUID, serviceBindingGUID, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.New())
	gomega.Expect(unbindResponse.Error).WithOffset(offset).NotTo(gomega.HaveOccurred())
	gomega.Expect(unbindResponse.StatusCode).WithOffset(offset).To(gomega.Equal(http.StatusOK))
}

func (h TestHelper) LastOperation(serviceInstanceGUID string, additionalOffset ...int) (result domain.LastOperation) {
	offset := computeOffset(1, additionalOffset)

	lastOperationResponse := h.Client().LastOperation(serviceInstanceGUID, uuid.New())
	gomega.Expect(lastOperationResponse.Error).WithOffset(offset).NotTo(gomega.HaveOccurred())
	gomega.Expect(lastOperationResponse.StatusCode).WithOffset(offset).To(gomega.Equal(http.StatusOK))
	gomega.Expect(json.Unmarshal(lastOperationResponse.ResponseBody, &result)).WithOffset(offset).NotTo(gomega.HaveOccurred())
	return result
}

func (h *TestHelper) LastOperationFinalState(serviceInstanceGUID string, additionalOffset ...int) domain.LastOperationState {
	offset := computeOffset(1, additionalOffset)

	start := time.Now()
	for {
		lastOperation := h.LastOperation(serviceInstanceGUID, offset+1)

		switch {
		case time.Since(start) > timeout:
			ginkgo.Fail(fmt.Sprintf("timed out waiting for last operation on service instance %q", serviceInstanceGUID), offset)
		case lastOperation.State == domain.Failed, lastOperation.State == domain.Succeeded:
			return lastOperation.State
		default:
			time.Sleep(pollingInterval)
		}
	}
}

func toJSONRawMessage(params []any, offset int) json.RawMessage {
	switch len(params) {
	case 0:
		return nil
	case 1:
	default:
		ginkgo.Fail("too many parameters passed", offset)
	}

	switch p := params[0].(type) {
	case nil:
		return nil
	case string:
		return json.RawMessage(p)
	case []byte:
		return p
	default:
		result, err := json.Marshal(p)
		gomega.Expect(offset, err).NotTo(gomega.HaveOccurred())
		return result
	}
}

func computeOffset(base int, additional []int) int {
	switch len(additional) {
	case 1:
		return base + additional[0]
	default:
		return base
	}
}
