package ccapi

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"
)

func UpgradeServiceInstance(req requester.Requester, guid, miVersion string) error {
	body := requestBody{
		MaintenanceInfoVersion: miVersion,
	}
	err := req.Patch(fmt.Sprintf("v3/service_instances/%s", guid), body)
	if err != nil {
		return fmt.Errorf("upgrade request error: %s", err)
	}

	var si ServiceInstance
	for {
		err = req.Get(fmt.Sprintf("v3/service_instances/%s", guid), &si)
		if err != nil {
			return fmt.Errorf("upgrade request error: %s", err)
		}

		if si.LastOperation.State == "failed" && si.LastOperation.Type == "update" {
			return fmt.Errorf("%s", si.LastOperation.Description)
		}

		if si.LastOperation.State != "in progress" || si.LastOperation.Type != "update" {
			return nil
		}
	}

	return nil
}

type requestBody struct {
	MaintenanceInfoVersion string `jsonry:"maintenance_info.version"`
}
