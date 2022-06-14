package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	apiToken := os.Getenv("CF_TOKEN")
	if apiToken == "" {
		panic("no token")
	}
	apiURL := os.Getenv("CF_API")
	if apiURL == "" {
		panic("no api url")
	}
	brokerName := os.Getenv("BROKER_NAME")
	if brokerName == "" {
		panic("no broker name")
	}

	// cf curl '/v3/service_plans?service_broker_names=csb-gblue' | jfq resources.guid

	// cf curl '/v3/service_instances?per_page=5000' | jfq 'resources[upgrade_available=true].{"guid":guid,"plan_guid":relationships.service_plan.data.guid}'
	//for each unique plan
	//get MI version: cf curl /v3/service_plans/a6e76697-0360-48b7-93de-2447591e6e39 | jfq maintenance_info.version
	//for each service instance:
	//trigger update: cf curl -X PATCH /v3/service_instances/a723ea6c-37cf-4685-bfc4-6b73aceed6fe -d '{"maintenance_info":{"version":"1.1.6"}}'
	//poll until complete: cf curl /v3/service_instances/a723ea6c-37cf-4685-bfc4-6b73aceed6fe | jfq last_operation.state

	plans := plansForBroker(apiURL, apiToken, brokerName)

	var planGUIDs []string
	planVersions := make(map[string]string)

	for _, p := range plans {
		planGUIDs = append(planGUIDs, p.GUID)
		planVersions[p.GUID] = p.MaintenanceInfo.Version
	}

	instances := serviceInstances(apiURL, apiToken, planGUIDs)

	for _, instance := range instances {
		if instance.UpgradeAvailable {
			newVersion := planVersions[instance.Relationships.ServicePlan.Data.GUID]
			upgrade(instance.GUID, newVersion)
		}
	}
}

type plan struct {
	GUID            string `json:"guid"`
	MaintenanceInfo struct {
		Version string `json:"version"`
	} `json:"maintenance_info"`
}

func plansForBroker(apiURL, apiToken, brokerName string) []plan {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v3/service_plans?per_page=5000&service_broker_names=%s", apiURL, brokerName), nil)
	if err != nil {
		panic(err)
	}
	request.Header.Add("Authorization", apiToken)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Minute,
	}

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var receiver struct {
		Resources []plan `json:"resources"`
	}
	if err := json.Unmarshal(data, &receiver); err != nil {
		panic(err)
	}

	return receiver.Resources
}

type serviceInstance struct {
	GUID             string `json:"guid"`
	UpgradeAvailable bool   `json:"upgrade_available"`
	Relationships    struct {
		ServicePlan struct {
			Data struct {
				GUID string `json:"guid"`
			} `json:"data"`
		} `json:"service_plan"`
	} `json:"relationships"`
}

func serviceInstances(apiURL, apiToken string, planGUIDs []string) []serviceInstance {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v3/service_instances?per_page=5000&service_plan_guids=%s", apiURL, strings.Join(planGUIDs, ",")), nil)
	if err != nil {
		panic(err)
	}
	request.Header.Add("Authorization", apiToken)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Minute,
	}

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var receiver struct {
		Resources []serviceInstance `json:"resources"`
	}
	if err := json.Unmarshal(data, &receiver); err != nil {
		panic(err)
	}

	return receiver.Resources
}

func upgrade(serviceInstanceGUID, newVersion string) {
	fmt.Printf("upgrading %s to %s\n", serviceInstanceGUID, newVersion)
}
