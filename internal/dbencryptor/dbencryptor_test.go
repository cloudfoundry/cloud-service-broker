package dbencryptor_test

import (
	"context"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/dbencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/compoundencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/gcmencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/noopencryptor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("EncryptDB", func() {
	const jsonSecret = `{"a":"secret"}`

	var (
		db                        *gorm.DB
		key                       [32]byte
		serviceInstanceDetails    models.ServiceInstanceDetails
		provisionRequestDetails   models.ProvisionRequestDetails
		serviceBindingCredentials models.ServiceBindingCredentials
		terraformDeployment       models.TerraformDeployment
		mapSecret                 map[string]interface{}
	)

	BeforeEach(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		db.Migrator().CreateTable(models.ServiceInstanceDetails{})
		db.Migrator().CreateTable(models.ServiceBindingCredentials{})
		db.Migrator().CreateTable(models.ProvisionRequestDetails{})
		db.Migrator().CreateTable(models.TerraformDeployment{})

		db_service.DbConnection = db
		models.SetEncryptor(noopencryptor.New())

		mapSecret = map[string]interface{}{"a": "secret"}

		copy(key[:], "one-key-here-with-32-bytes-in-it")

		serviceInstanceDetails = models.ServiceInstanceDetails{}
		serviceInstanceDetails.SetOtherDetails(mapSecret)
		Expect(db_service.CreateServiceInstanceDetails(context.TODO(), &serviceInstanceDetails)).NotTo(HaveOccurred())

		serviceBindingCredentials = models.ServiceBindingCredentials{}
		serviceBindingCredentials.SetOtherDetails(mapSecret)
		Expect(db_service.CreateServiceBindingCredentials(context.TODO(), &serviceBindingCredentials)).NotTo(HaveOccurred())

		provisionRequestDetails = models.ProvisionRequestDetails{ServiceInstanceId: uuid.New()}
		Expect(provisionRequestDetails.SetRequestDetails([]byte(jsonSecret))).NotTo(HaveOccurred())
		Expect(db_service.CreateProvisionRequestDetails(context.TODO(), &provisionRequestDetails)).NotTo(HaveOccurred())

		terraformDeployment = models.TerraformDeployment{}
		terraformDeployment.SetWorkspace(jsonSecret)
		Expect(db_service.CreateTerraformDeployment(context.TODO(), &terraformDeployment)).NotTo(HaveOccurred())
	})

	persistedServiceInstanceDetails := func() string {
		record := models.ServiceInstanceDetails{}
		Expect(db.First(&record).Error).NotTo(HaveOccurred())
		return record.OtherDetails
	}

	persistedRequestDetails := func() string {
		record := models.ProvisionRequestDetails{}
		Expect(db.First(&record).Error).NotTo(HaveOccurred())
		return record.RequestDetails
	}

	persistedServiceBindingDetails := func() string {
		record := models.ServiceBindingCredentials{}
		Expect(db.First(&record).Error).NotTo(HaveOccurred())
		return record.OtherDetails
	}

	persistedTerraformWorkspace := func() string {
		record := models.TerraformDeployment{}
		Expect(db.First(&record).Error).NotTo(HaveOccurred())
		return record.Workspace
	}

	It("encrypts the database", func() {
		models.SetEncryptor(compoundencryptor.New(
			gcmencryptor.New(key),
			noopencryptor.New(),
		))

		Expect(persistedServiceInstanceDetails()).To(Equal(jsonSecret))
		Expect(persistedRequestDetails()).To(Equal(jsonSecret))
		Expect(persistedServiceBindingDetails()).To(Equal(jsonSecret))
		Expect(persistedTerraformWorkspace()).To(Equal(jsonSecret))

		By("running the encryption")
		Expect(dbencryptor.EncryptDB(db)).NotTo(HaveOccurred())

		Expect(persistedServiceInstanceDetails()).NotTo(Equal(jsonSecret))
		Expect(persistedRequestDetails()).NotTo(Equal(jsonSecret))
		Expect(persistedServiceBindingDetails()).NotTo(Equal(jsonSecret))
		Expect(persistedTerraformWorkspace()).NotTo(Equal(jsonSecret))
	})
})
