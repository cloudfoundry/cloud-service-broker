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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("ReencryptDB", func() {
	const jsonSecret = `{"a":"secret"}`

	var (
		db                     *gorm.DB
		key                    [32]byte
		serviceInstanceDetails models.ServiceInstanceDetails
		terraformDeployment    models.TerraformDeployment
		mapSecret              map[string]interface{}
	)

	persistedServiceInstanceDetails := func() []byte {
		record := models.ServiceInstanceDetails{}
		Expect(db.First(&record).Error).NotTo(HaveOccurred())
		return record.OtherDetails
	}

	persistedTerraformWorkspace := func() []byte {
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
			db.Migrator().CreateTable(models.TerraformDeployment{})

			db_service.DbConnection = db
			models.SetEncryptor(noopencryptor.New())

			mapSecret = map[string]interface{}{"a": "secret"}

			copy(key[:], "one-key-here-with-32-bytes-in-it")

			serviceInstanceDetails = models.ServiceInstanceDetails{}
			serviceInstanceDetails.SetOtherDetails(mapSecret)
			Expect(db_service.CreateServiceInstanceDetails(context.TODO(), &serviceInstanceDetails)).NotTo(HaveOccurred())

			terraformDeployment = models.TerraformDeployment{}
			terraformDeployment.SetWorkspace(jsonSecret)
			Expect(db_service.CreateTerraformDeployment(context.TODO(), &terraformDeployment)).NotTo(HaveOccurred())
		})

		It("encrypts the database", func() {
			models.SetEncryptor(compoundencryptor.New(
				gcmencryptor.New(key),
				noopencryptor.New(),
			))

			Expect(persistedServiceInstanceDetails()).To(Equal([]byte(jsonSecret)))
			Expect(persistedTerraformWorkspace()).To(Equal([]byte(jsonSecret)))

			By("running the encryption")
			Expect(dbrotator.ReencryptDB(db)).NotTo(HaveOccurred())

			Expect(persistedServiceInstanceDetails()).NotTo(Equal([]byte(jsonSecret)))
			Expect(persistedTerraformWorkspace()).NotTo(Equal([]byte(jsonSecret)))
		})
	})

	When("db was encrypted", func() {
		BeforeEach(func() {
			var err error
			db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			Expect(err).NotTo(HaveOccurred())
			db.Migrator().CreateTable(models.ServiceInstanceDetails{})
			db.Migrator().CreateTable(models.TerraformDeployment{})

			db_service.DbConnection = db
			copy(key[:], "one-key-here-with-32-bytes-in-it")
			models.SetEncryptor(gcmencryptor.New(key))

			mapSecret = map[string]interface{}{"a": "secret"}

			serviceInstanceDetails = models.ServiceInstanceDetails{}
			serviceInstanceDetails.SetOtherDetails(mapSecret)
			Expect(db_service.CreateServiceInstanceDetails(context.TODO(), &serviceInstanceDetails)).NotTo(HaveOccurred())

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
			Expect(persistedTerraformWorkspace()).NotTo(Equal(jsonSecret))

			By("running the encryption")
			Expect(dbrotator.ReencryptDB(db)).NotTo(HaveOccurred())

			Expect(persistedServiceInstanceDetails()).To(Equal([]byte(jsonSecret)))
			Expect(persistedTerraformWorkspace()).To(Equal([]byte(jsonSecret)))
		})
	})

	When("db encryption key changes", func() {
		var newKey [32]byte

		BeforeEach(func() {
			var err error
			db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			Expect(err).NotTo(HaveOccurred())
			db.Migrator().CreateTable(models.ServiceInstanceDetails{})
			db.Migrator().CreateTable(models.TerraformDeployment{})

			db_service.DbConnection = db
			copy(key[:], "one-key-here-with-32-bytes-in-it")
			models.SetEncryptor(gcmencryptor.New(key))

			mapSecret = map[string]interface{}{"a": "secret"}

			serviceInstanceDetails = models.ServiceInstanceDetails{ID: "1"}
			serviceInstanceDetails.SetOtherDetails(mapSecret)
			Expect(db_service.SaveServiceInstanceDetails(context.TODO(), &serviceInstanceDetails)).NotTo(HaveOccurred())

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
			firstEncryptionPersistedTerraformWorkspace := persistedTerraformWorkspace()
			Expect(firstEncryptionPersistedTerraformWorkspace).NotTo(Equal(jsonSecret))

			By("running the encryption")
			Expect(dbrotator.ReencryptDB(db)).NotTo(HaveOccurred())

			Expect(persistedServiceInstanceDetails()).NotTo(Equal(jsonSecret))
			Expect(persistedServiceInstanceDetails()).NotTo(Equal(firstEncryptionPersistedServiceInstanceDetails))
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
				record.OtherDetails = []byte("something-that-cannot-be-decrypted-with-provided-decryptors")
				Expect(db.Save(&record).Error).NotTo(HaveOccurred())

				By("running the encryption")
				Expect(dbrotator.ReencryptDB(db)).To(MatchError("error reencrypting: cipher: message authentication failed"))
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
				record.Workspace = []byte("something-that-cannot-be-decrypted-with-provided-decryptors")
				Expect(db.Save(&record).Error).NotTo(HaveOccurred())

				By("running the encryption")
				Expect(dbrotator.ReencryptDB(db)).To(MatchError("error reencrypting: cipher: message authentication failed"))
			})
		})
	})
})
