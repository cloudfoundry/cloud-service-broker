package broker_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"golang.org/x/net/context"

	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker"
	"github.com/cloudfoundry/cloud-service-broker/brokerapi/broker/brokerfakes"
	pkgBroker "github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/utils"
)

var _ = Describe("Services", func() {
	var (
		serviceBroker *broker.ServiceBroker
	)

	BeforeEach(func() {
		brokerConfig := &broker.BrokerConfig{
			Registry: pkgBroker.BrokerRegistry{
				"first-service": &pkgBroker.ServiceDefinition{
					ID:                  "first-service-id",
					Name:                "first-service",
					ProviderDisplayName: "company-name-1",
					Plans: []pkgBroker.ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID:   "plan-1",
								Name: "test-plan-1",
							},
						},
						{
							ServicePlan: domain.ServicePlan{
								ID:   "plan-2",
								Name: "test-plan-2",
							},
						},
					},
				},
				"second-service": &pkgBroker.ServiceDefinition{
					ID:                  "second-service-id",
					Name:                "second-service",
					ProviderDisplayName: "company-name-2",
					Plans: []pkgBroker.ServicePlan{
						{
							ServicePlan: domain.ServicePlan{
								ID:   "plan-3",
								Name: "test-plan-3",
							},
						},
					},
				},
			},
		}

		serviceBroker = must(broker.New(brokerConfig, &brokerfakes.FakeStorage{}, utils.NewLogger("brokers-test")))
	})

	Describe("getting services", func() {
		It("should return list of service offerings", func() {
			servicesList, err := serviceBroker.Services(context.TODO())

			Expect(err).ToNot(HaveOccurred())
			Expect(len(servicesList)).To(Equal(2))

			Expect(servicesList[0].ID).To(Equal("first-service-id"))
			Expect(servicesList[0].Name).To(Equal("first-service"))
			Expect(servicesList[0].Metadata.ProviderDisplayName).To(Equal("company-name-1"))
			Expect(len(servicesList[0].Plans)).To(Equal(2))
			Expect(servicesList[0].Plans[0].ID).To(Equal("plan-1"))
			Expect(servicesList[0].Plans[0].Name).To(Equal("test-plan-1"))
			Expect(servicesList[0].Plans[1].ID).To(Equal("plan-2"))
			Expect(servicesList[0].Plans[1].Name).To(Equal("test-plan-2"))

			Expect(servicesList[1].ID).To(Equal("second-service-id"))
			Expect(servicesList[1].Name).To(Equal("second-service"))
			Expect(servicesList[1].Metadata.ProviderDisplayName).To(Equal("company-name-2"))
			Expect(len(servicesList[1].Plans)).To(Equal(1))
			Expect(servicesList[1].Plans[0].ID).To(Equal("plan-3"))
			Expect(servicesList[1].Plans[0].Name).To(Equal("test-plan-3"))
		})
	})
})
