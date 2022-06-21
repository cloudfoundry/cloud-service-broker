package ccapi

import (
	"fmt"

	"code.cloudfoundry.org/jsonry"
	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"
)

func GetServicePlans(r requester.Requester, brokerName string) ([]Plan, error) {
	var brokerPlans servicePlans
	if err := r.Get(fmt.Sprintf("v3/service_plans?per_page=5000&service_broker_names=%s", brokerName), &brokerPlans); err != nil {
		return nil, fmt.Errorf("error getting service plans: %s", err)
	}
	return brokerPlans.Plans, nil
}

type servicePlans struct {
	Plans []Plan `json:"resources"`
}

type Plan struct {
	GUID                   string `json:"guid"`
	MaintenanceInfoVersion string `jsonry:"maintenance_info.version"`
}

func (p *Plan) UnmarshalJSON(b []byte) error {
	return jsonry.Unmarshal(b, p)
}
