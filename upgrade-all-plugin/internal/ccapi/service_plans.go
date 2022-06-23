package ccapi

import (
	"fmt"

	"code.cloudfoundry.org/jsonry"
)

type servicePlans struct {
	Plans []Plan `json:"resources"`
}

type Plan struct {
	GUID                   string `json:"guid"`
	MaintenanceInfoVersion string `jsonry:"maintenance_info.version"`
}

func (c CCAPI) GetServicePlans(brokerName string) ([]Plan, error) {
	var brokerPlans servicePlans
	if err := c.requester.Get(fmt.Sprintf("v3/service_plans?per_page=5000&service_broker_names=%s", brokerName), &brokerPlans); err != nil {
		return nil, fmt.Errorf("error getting service plans: %s", err)
	}
	return brokerPlans.Plans, nil
}

func (p *Plan) UnmarshalJSON(b []byte) error {
	return jsonry.Unmarshal(b, p)
}
