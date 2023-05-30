// Package displaycatalog is used to print out the key data of the catalog in the logs without
// printing out binary images etc... which also form part of the catalog
package displaycatalog

import "github.com/pivotal-cf/brokerapi/v10/domain"

func DisplayCatalog(services []domain.Service) []any {
	return mapSlice(services, func(service domain.Service) any {
		return map[string]any{
			"name": service.Name,
			"id":   service.ID,
			"tags": service.Tags,
			"plans": mapSlice(service.Plans, func(plan domain.ServicePlan) any {
				return map[string]any{
					"name": plan.Name,
					"id":   plan.ID,
				}
			}),
		}
	})
}

func mapSlice[A any](s []A, cb func(A) any) (result []any) {
	for _, e := range s {
		result = append(result, cb(e))
	}
	return
}
