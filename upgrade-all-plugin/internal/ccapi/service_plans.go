package ccapi

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/upgrade-all-plugin/internal/requester"
)

func GetServicePlans(r requester.Requester, brokerName string) ([]Plan, error) {
	var brokerPlans ServicePlans
	if err := r.Get(fmt.Sprintf("v3/service_plans?per_page=5000&service_broker_names=%s", brokerName), &brokerPlans); err != nil {
		return nil, fmt.Errorf("error getting service plans: %s", err)
	}
	return brokerPlans.Plans, nil
}

type Plan struct {
	GUID            string `json:"guid"`
	MaintenanceInfo struct {
		Version string `json:"version"`
	} `json:"maintenance_info"`
}

type ServicePlans struct {
	Plans []Plan `json:"resources"`
}
