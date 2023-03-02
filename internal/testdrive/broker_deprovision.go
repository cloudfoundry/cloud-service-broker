package testdrive

import (
	"fmt"
	"net/http"

	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v9/domain"
)

func (b *Broker) Deprovision(s ServiceInstance) error {
	deprovisionResponse := b.Client.Deprovision(s.GUID, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.New())
	switch {
	case deprovisionResponse.Error != nil:
		return deprovisionResponse.Error
	case deprovisionResponse.StatusCode != http.StatusAccepted:
		return &UnexpectedStatusError{StatusCode: deprovisionResponse.StatusCode, ResponseBody: deprovisionResponse.ResponseBody}
	}

	state, err := b.LastOperationFinalState(s.GUID)
	switch {
	case err != nil:
		return err
	case state != domain.Succeeded:
		return fmt.Errorf("deprovison failed with state: %s", state)
	}

	return nil
}
