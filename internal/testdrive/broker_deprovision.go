package testdrive

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/pivotal-cf/brokerapi/v11/domain"
)

func (b *Broker) Deprovision(s ServiceInstance) error {
	deprovisionResponse := b.Client.Deprovision(s.GUID, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.NewString())
	switch {
	case deprovisionResponse.Error != nil:
		return deprovisionResponse.Error
	case deprovisionResponse.StatusCode == http.StatusOK: // ok - synchronous
		return nil
	case deprovisionResponse.StatusCode == http.StatusAccepted: // ok - asynchronous - poll last operation
	default:
		return &UnexpectedStatusError{StatusCode: deprovisionResponse.StatusCode, ResponseBody: deprovisionResponse.ResponseBody}
	}

	lastOperation, err := b.LastOperationFinalValue(s.GUID)
	switch {
	case err != nil:
		return err
	case lastOperation.State != domain.Succeeded:
		return fmt.Errorf("deprovison failed with state: %s and error: %s", lastOperation.State, lastOperation.Description)
	}

	return nil
}
