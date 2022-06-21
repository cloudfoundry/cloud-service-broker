package ccapi

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"
)

func UpgradeServiceInstance(req requester.Requester, guid, miVersion string) error {
	body := requestBody{
		MaintenanceInfoVersion: miVersion,
	}
	req.Patch(fmt.Sprintf("v3/service_instances/%s", guid), body)

	// Poll upgrade request and block till complete

	return nil
}

type requestBody struct {
	MaintenanceInfoVersion string `jsonry:"maintenance_info.version"`
}
