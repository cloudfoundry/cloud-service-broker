package testdrive

import (
	"net/http"

	"github.com/pborman/uuid"
)

func (b *Broker) DeleteBinding(s ServiceInstance, serviceBindingGUID string) error {
	unbindResponse := b.Client.Unbind(s.GUID, serviceBindingGUID, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.New())
	switch {
	case unbindResponse.Error != nil:
		return unbindResponse.Error
	case unbindResponse.StatusCode != http.StatusOK:
		return &UnexpectedStatusError{StatusCode: unbindResponse.StatusCode, ResponseBody: unbindResponse.ResponseBody}
	default:
		return nil
	}
}
