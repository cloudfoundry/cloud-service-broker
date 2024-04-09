package displaycatalog_test

import (
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/displaycatalog"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v11/domain"
)

var _ = Describe("DisplayCatalog", func() {
	It("generate the right details to display", func() {
		catalog := []domain.Service{
			{
				ID:   "fake-service-id-1",
				Name: "fake-service-name-1",
				Tags: []string{"fake-tag-1-1", "fake-tag-1-2"},
				Plans: []domain.ServicePlan{
					{
						ID:   "fake-plan-id-1",
						Name: "fake-plan-name-1",
					},
					{
						ID:   "fake-plan-id-2",
						Name: "fake-plan-name-2",
					},
				},
			},
			{
				ID:   "fake-service-id-2",
				Name: "fake-service-name-2",
				Tags: []string{"fake-tag-2-1", "fake-tag-2-2"},
				Plans: []domain.ServicePlan{
					{
						ID:   "fake-plan-id-3",
						Name: "fake-plan-name-3",
					},
					{
						ID:   "fake-plan-id-4",
						Name: "fake-plan-name-4",
					},
				},
			},
		}

		Expect(displaycatalog.DisplayCatalog(catalog)).To(Equal([]any{
			map[string]any{
				"name": "fake-service-name-1",
				"id":   "fake-service-id-1",
				"tags": []string{"fake-tag-1-1", "fake-tag-1-2"},
				"plans": []any{
					map[string]any{
						"name": "fake-plan-name-1",
						"id":   "fake-plan-id-1",
					},
					map[string]any{
						"name": "fake-plan-name-2",
						"id":   "fake-plan-id-2",
					},
				},
			},
			map[string]any{
				"name": "fake-service-name-2",
				"id":   "fake-service-id-2",
				"tags": []string{"fake-tag-2-1", "fake-tag-2-2"},
				"plans": []any{
					map[string]any{
						"name": "fake-plan-name-3",
						"id":   "fake-plan-id-3",
					},
					map[string]any{
						"name": "fake-plan-name-4",
						"id":   "fake-plan-id-4",
					},
				},
			},
		}))
	})
})
