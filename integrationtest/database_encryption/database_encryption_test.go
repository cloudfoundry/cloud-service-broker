package database_encryption_test

import (
	"context"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/cloud-service-broker/brokerapi/brokers"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker/brokerfakes"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Database Encryption Integration", func() {
	var (
		tmpDBFile           string
		fakeService         broker.ServiceDefinition
		fakeServiceProvider broker.ServiceProvider
		serviceBroker       *brokers.ServiceBroker
		db                  *gorm.DB
	)

	BeforeEach(func() {
		fakeServiceProvider = &brokerfakes.FakeServiceProvider{}
		fakeService = broker.ServiceDefinition{
			Id:   uuid.New(),
			Name: "fake-service-name",
			Plans: []broker.ServicePlan{
				{
					ServicePlan: domain.ServicePlan{
						ID:   uuid.New(),
						Name: "fake-plan-name",
					},
					ServiceProperties:  nil,
					ProvisionOverrides: nil,
					BindOverrides:      nil,
				},
			},
			ProviderBuilder: func(lager.Logger) broker.ServiceProvider {
				return fakeServiceProvider
			},
		}

		err := os.Setenv("DB_TYPE", "sqlite3")
		Expect(err).NotTo(HaveOccurred())
		fh, err := os.CreateTemp("", "*-csb-test")
		Expect(err).NotTo(HaveOccurred())
		tmpDBFile = fh.Name()
		fh.Close()
		err = os.Setenv("DB_PATH", tmpDBFile)
		Expect(err).NotTo(HaveOccurred())

		logger := utils.NewLogger("csb-test")
		db = db_service.New(logger)

		registry := broker.BrokerRegistry{}
		registry.Register(&fakeService)
		cfg := brokers.BrokerConfig{
			Registry:  registry,
			Credstore: nil,
		}

		serviceBroker, err = brokers.New(&cfg, logger)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.Remove(tmpDBFile)
		Expect(err).NotTo(HaveOccurred())
	})

	It("stores the provision request details in plaintext", func() {
		const requestDetails = `{"foo":"bar"}`

		instanceID := uuid.New()
		details := domain.ProvisionDetails{
			ServiceID:        fakeService.Id,
			PlanID:           fakeService.Plans[0].ID,
			OrganizationGUID: uuid.New(),
			SpaceGUID:        uuid.New(),
			RawParameters:    []byte(requestDetails),
		}
		const async = false
		serviceBroker.Provision(context.TODO(), instanceID, details, async)

		record := models.ProvisionRequestDetails{}
		err := db.Where("service_instance_id = ?", instanceID).First(&record).Error
		Expect(err).NotTo(HaveOccurred())

		Expect(record.RequestDetails).To(Equal(requestDetails))
	})
})
