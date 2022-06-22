package ccapi

import (
	"fmt"
	"time"
)

func (c CCAPI) UpgradeServiceInstance(guid, miVersion string) error {
	fmt.Printf("Upgrading instance: %s with version: %s\n", guid, miVersion)

	var body struct {
		MaintenanceInfo struct {
			Version string `json:"version"`
		} `json:"maintenance_info"`
	}
	body.MaintenanceInfo.Version = miVersion

	err := c.requester.Patch(fmt.Sprintf("v3/service_instances/%s", guid), body)
	if err != nil {
		return fmt.Errorf("upgrade request error: %s", err)
	}

	var si ServiceInstance
	for {
		fmt.Printf("Polling instance: %v\n", guid)

		err = c.requester.Get(fmt.Sprintf("v3/service_instances/%s", guid), &si)
		if err != nil {
			return fmt.Errorf("upgrade request error: %s", err)
		}

		if si.LastOperation.State == "failed" && si.LastOperation.Type == "update" {
			return fmt.Errorf("%s", si.LastOperation.Description)
		}

		if si.LastOperation.State != "in progress" || si.LastOperation.Type != "update" {
			return nil
		}

		time.Sleep(time.Second)
	}

	return nil
}
