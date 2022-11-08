package testdrive

import (
	"fmt"
	"net/http"

	"github.com/pborman/uuid"
)

func (b *Broker) DeleteBinding(s ServiceInstance, serviceBindingGUID string) error {
	unbindResponse := b.Client.Unbind(s.GUID, serviceBindingGUID, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.New())
	switch {
	case unbindResponse.Error != nil:
		return unbindResponse.Error
	case unbindResponse.StatusCode != http.StatusOK:
		return fmt.Errorf("unexpected status code %d: %s", unbindResponse.StatusCode, unbindResponse.ResponseBody)
	default:
		return nil
	}
}
