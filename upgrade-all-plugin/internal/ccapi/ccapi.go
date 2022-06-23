package ccapi

import "github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"

type CCAPI struct {
	requester requester.Requester
}

func NewCCAPI(req requester.Requester) CCAPI {
	return CCAPI{
		requester: req,
	}
}
