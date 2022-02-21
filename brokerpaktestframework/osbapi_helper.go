package brokerpaktestframework

import (
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)

func FindService(catalog *apiresponses.CatalogResponse, s string) domain.Service {
	for _, service := range catalog.Services {
		if service.Name == s {
			return service
		}
	}
	return domain.Service{}
}

func FindServicePlan(catalog *apiresponses.CatalogResponse, serviceName, servicePlan string) domain.ServicePlan {
	for _, service := range catalog.Services {
		if service.Name == serviceName {
			for _, plan := range service.Plans {
				if plan.Name == servicePlan {
					return plan
				}
			}
		}
	}
	return domain.ServicePlan{}
}

func FindServicePlanGUIDs(catalog *apiresponses.CatalogResponse, serviceName, planName string) (string, string) {
	service := FindService(catalog, serviceName)
	for _, plan := range service.Plans {
		if plan.Name == planName {
			return service.ID, plan.ID
		}
	}
	return "", ""
}
