package ccapi

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/jsonry"
)

type ServiceInstance struct {
	GUID             string `json:"guid"`
	UpgradeAvailable bool   `json:"upgrade_available"`
	PlanGUID         string `jsonry:"relationships.service_plan.data.guid"`
	LastOperation    struct {
		Type        string `json:"type"`
		State       string `json:"state"`
		Description string `json:"description"`
	} `json:"last_operation"`
}

type serviceInstances struct {
	Instances []ServiceInstance `json:"resources"`
}

func (c CCAPI) GetServiceInstances(planGUIDs []string) ([]ServiceInstance, error) {
	if len(planGUIDs) == 0 {
		return nil, fmt.Errorf("no service_plan_guids specified")
	}

	var si serviceInstances
	if err := c.requester.Get(fmt.Sprintf("v3/service_instances?per_page=5000&service_plan_guids=%s", strings.Join(planGUIDs, ",")), &si); err != nil {
		return nil, fmt.Errorf("error getting service instances: %s", err)
	}
	return si.Instances, nil
}

func (s *ServiceInstance) UnmarshalJSON(b []byte) error {
	return jsonry.Unmarshal(b, s)
}
