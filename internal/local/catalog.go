package local

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/pkg/client"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v9/domain"
)

func catalog(clnt *client.Client) []domain.Service {
	catalogResponse := clnt.Catalog(uuid.New())
	switch {
	case catalogResponse.Error != nil:
		log.Fatal(catalogResponse.Error)
	case catalogResponse.StatusCode != http.StatusOK:
		log.Fatalf("bad catalog response: %d", catalogResponse.StatusCode)
	}

	var c struct {
		Services []domain.Service `json:"services"`
	}
	if err := json.Unmarshal(catalogResponse.ResponseBody, &c); err != nil {
		log.Fatal(err)
	}

	return c.Services
}

func lookupServiceIDsByName(clnt *client.Client, serviceName, planName string) (string, string) {
	for _, s := range catalog(clnt) {
		if s.Name == serviceName {
			for _, p := range s.Plans {
				if p.Name == planName {
					return s.ID, p.ID
				}
			}
			panic(fmt.Sprintf("could not find plan %q in service %q", planName, serviceName))
		}
	}
	panic(fmt.Sprintf("could not find service %q in catalog", serviceName))
}

func lookupPlanIDByName(clnt *client.Client, serviceID, planName string) string {
	for _, s := range catalog(clnt) {
		if s.ID == serviceID {
			for _, p := range s.Plans {
				if p.Name == planName {
					return p.ID
				}
			}
			log.Fatalf("could not find plan %q in service %q", planName, serviceID)
		}
	}
	panic(fmt.Sprintf("could not find service %q in catalog", serviceID))
}

func lookupServiceNamesByID(clnt *client.Client, serviceID, planID string) (string, string) {
	for _, s := range catalog(clnt) {
		if s.ID == serviceID {
			for _, p := range s.Plans {
				if p.ID == planID {
					return s.Name, p.Name
				}
			}
			panic(fmt.Sprintf("could not find plan %q in service %q", planID, serviceID))
		}
	}
	panic(fmt.Sprintf("could not find service %q in catalog", serviceID))
}
