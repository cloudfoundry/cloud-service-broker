package dbrotator_test

import (
	"context"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/compoundencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/dbrotator"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/gcmencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/noopencryptor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("ReencryptDB", func() {
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

	When("db was not encrypted", func() {
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
			Expect(dbrotator.ReencryptDB(db)).NotTo(HaveOccurred())

			Expect(persistedServiceInstanceDetails()).NotTo(Equal(jsonSecret))
			Expect(persistedRequestDetails()).NotTo(Equal(jsonSecret))
			Expect(persistedServiceBindingDetails()).NotTo(Equal(jsonSecret))
			Expect(persistedTerraformWorkspace()).NotTo(Equal(jsonSecret))
		})
	})

	When("db was encrypted", func() {
		BeforeEach(func() {
			var err error
			db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			Expect(err).NotTo(HaveOccurred())
			db.Migrator().CreateTable(models.ServiceInstanceDetails{})
			db.Migrator().CreateTable(models.ServiceBindingCredentials{})
			db.Migrator().CreateTable(models.ProvisionRequestDetails{})
			db.Migrator().CreateTable(models.TerraformDeployment{})

			db_service.DbConnection = db
			copy(key[:], "one-key-here-with-32-bytes-in-it")
			models.SetEncryptor(gcmencryptor.New(key))

			mapSecret = map[string]interface{}{"a": "secret"}

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

		It("deencrypts the database", func() {
			models.SetEncryptor(compoundencryptor.New(
				noopencryptor.New(),
				gcmencryptor.New(key),
			))

			Expect(persistedServiceInstanceDetails()).NotTo(Equal(jsonSecret))
			Expect(persistedRequestDetails()).NotTo(Equal(jsonSecret))
			Expect(persistedServiceBindingDetails()).NotTo(Equal(jsonSecret))
			Expect(persistedTerraformWorkspace()).NotTo(Equal(jsonSecret))

			By("running the encryption")
			Expect(dbrotator.ReencryptDB(db)).NotTo(HaveOccurred())

			Expect(persistedServiceInstanceDetails()).To(Equal(jsonSecret))
			Expect(persistedRequestDetails()).To(Equal(jsonSecret))
			Expect(persistedServiceBindingDetails()).To(Equal(jsonSecret))
			Expect(persistedTerraformWorkspace()).To(Equal(jsonSecret))
		})
	})

	When("db encryption key changes", func() {
		var newKey [32]byte

		BeforeEach(func() {
			var err error
			db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			Expect(err).NotTo(HaveOccurred())
			db.Migrator().CreateTable(models.ServiceInstanceDetails{})
			db.Migrator().CreateTable(models.ServiceBindingCredentials{})
			db.Migrator().CreateTable(models.ProvisionRequestDetails{})
			db.Migrator().CreateTable(models.TerraformDeployment{})

			db_service.DbConnection = db
			copy(key[:], "one-key-here-with-32-bytes-in-it")
			models.SetEncryptor(gcmencryptor.New(key))

			mapSecret = map[string]interface{}{"a": "secret"}

			serviceInstanceDetails = models.ServiceInstanceDetails{ID: "1"}
			serviceInstanceDetails.SetOtherDetails(mapSecret)
			Expect(db_service.SaveServiceInstanceDetails(context.TODO(), &serviceInstanceDetails)).NotTo(HaveOccurred())

			serviceBindingCredentials = models.ServiceBindingCredentials{}
			serviceBindingCredentials.SetOtherDetails(mapSecret)
			Expect(db_service.CreateServiceBindingCredentials(context.TODO(), &serviceBindingCredentials)).NotTo(HaveOccurred())

			provisionRequestDetails = models.ProvisionRequestDetails{ServiceInstanceId: uuid.New()}
			Expect(provisionRequestDetails.SetRequestDetails([]byte(jsonSecret))).NotTo(HaveOccurred())
			Expect(db_service.CreateProvisionRequestDetails(context.TODO(), &provisionRequestDetails)).NotTo(HaveOccurred())

			terraformDeployment = models.TerraformDeployment{ID: "1"}
			terraformDeployment.SetWorkspace(jsonSecret)
			Expect(db_service.CreateTerraformDeployment(context.TODO(), &terraformDeployment)).NotTo(HaveOccurred())

			copy(newKey[:], "another-key-here-with-32-bytes-in-it")
		})

		It("re-encrypts the database with new key", func() {
			newEncryptor := gcmencryptor.New(newKey)

			models.SetEncryptor(compoundencryptor.New(
				newEncryptor,
				gcmencryptor.New(key),
				newEncryptor,
			))

			firstEncryptionPersistedServiceInstanceDetails := persistedServiceInstanceDetails()
			Expect(firstEncryptionPersistedServiceInstanceDetails).NotTo(Equal(jsonSecret))
			firstEncryptionPersistedRequestDetails := persistedRequestDetails()
			Expect(firstEncryptionPersistedRequestDetails).NotTo(Equal(jsonSecret))
			firstEncryptionPersistedServiceBindingDetails := persistedServiceBindingDetails()
			Expect(firstEncryptionPersistedServiceBindingDetails).NotTo(Equal(jsonSecret))
			firstEncryptionPersistedTerraformWorkspace := persistedTerraformWorkspace()
			Expect(firstEncryptionPersistedTerraformWorkspace).NotTo(Equal(jsonSecret))

			By("running the encryption")
			Expect(dbrotator.ReencryptDB(db)).NotTo(HaveOccurred())

			Expect(persistedServiceInstanceDetails()).NotTo(Equal(jsonSecret))
			Expect(persistedServiceInstanceDetails()).NotTo(Equal(firstEncryptionPersistedServiceInstanceDetails))
			Expect(persistedRequestDetails()).NotTo(Equal(jsonSecret))
			Expect(persistedRequestDetails()).NotTo(Equal(firstEncryptionPersistedRequestDetails))
			Expect(persistedServiceBindingDetails()).NotTo(Equal(jsonSecret))
			Expect(persistedServiceBindingDetails()).NotTo(Equal(firstEncryptionPersistedServiceBindingDetails))
			Expect(persistedTerraformWorkspace()).NotTo(Equal(jsonSecret))
			Expect(persistedTerraformWorkspace()).NotTo(Equal(firstEncryptionPersistedTerraformWorkspace))
		})

		Context("ServiceInstanceDetails", func() {
			It("returns error when decryption fails", func() {
				newEncryptor := gcmencryptor.New(newKey)

				models.SetEncryptor(compoundencryptor.New(
					newEncryptor,
					gcmencryptor.New(key),
					newEncryptor,
				))

				db_service.DbConnection = db

				record := models.ServiceInstanceDetails{}
				Expect(db.First(&record).Error).NotTo(HaveOccurred())
				record.OtherDetails = "something-that-cannot-be-decrypted-with-provided-decryptors"
				Expect(db.Save(&record).Error).NotTo(HaveOccurred())

				By("running the encryption")
				Expect(dbrotator.ReencryptDB(db)).To(MatchError("error reencrypting: illegal base64 data at input byte 9"))
			})
		})

		Context("ProvisionRequestDetails", func() {
			It("returns error when decryption fails", func() {
				newEncryptor := gcmencryptor.New(newKey)

				models.SetEncryptor(compoundencryptor.New(
					newEncryptor,
					gcmencryptor.New(key),
					newEncryptor,
				))

				db_service.DbConnection = db

				record := models.ProvisionRequestDetails{}
				Expect(db.First(&record).Error).NotTo(HaveOccurred())
				record.RequestDetails = "something-that-cannot-be-decrypted-with-provided-decryptors"
				Expect(db.Save(&record).Error).NotTo(HaveOccurred())

				By("running the encryption")
				Expect(dbrotator.ReencryptDB(db)).To(MatchError("error reencrypting: illegal base64 data at input byte 9"))
			})
		})

		Context("ServiceBindingDetails", func() {
			It("returns error when decryption fails", func() {
				newEncryptor := gcmencryptor.New(newKey)

				models.SetEncryptor(compoundencryptor.New(
					newEncryptor,
					gcmencryptor.New(key),
					newEncryptor,
				))

				db_service.DbConnection = db

				record := models.ServiceBindingCredentials{}
				Expect(db.First(&record).Error).NotTo(HaveOccurred())
				record.OtherDetails = "something-that-cannot-be-decrypted-with-provided-decryptors"
				Expect(db.Save(&record).Error).NotTo(HaveOccurred())

				By("running the encryption")
				Expect(dbrotator.ReencryptDB(db)).To(MatchError("error reencrypting: illegal base64 data at input byte 9"))
			})
		})

		Context("TerraformWorkspace", func() {
			It("returns error when decryption fails", func() {
				newEncryptor := gcmencryptor.New(newKey)

				models.SetEncryptor(compoundencryptor.New(
					newEncryptor,
					gcmencryptor.New(key),
					newEncryptor,
				))

				db_service.DbConnection = db

				record := models.TerraformDeployment{}
				Expect(db.First(&record).Error).NotTo(HaveOccurred())
				record.Workspace = "something-that-cannot-be-decrypted-with-provided-decryptors"
				Expect(db.Save(&record).Error).NotTo(HaveOccurred())

				By("running the encryption")
				Expect(dbrotator.ReencryptDB(db)).To(MatchError("error reencrypting: illegal base64 data at input byte 9"))
			})
		})
	})
})
