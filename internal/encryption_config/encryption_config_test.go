package encryption_config_test

import (
	"context"
	"os"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption_config"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("Encryption Config", func() {
	AfterEach(func() {
		viper.Reset()
	})

	Describe("GetEncryptionKey", func() {
		Describe("encryption is not enabled", func() {
			It("should return empty key", func() {
				key, err := encryption_config.GetEncryptionKey(false, "")
				Expect(err).ToNot(HaveOccurred())
				Expect(key).To(BeEmpty())
			})

			It("should return error when a primary password is also provided", func() {
				rawPasswords := "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"foo-foo\",\"primary\":true}]"

				_, err := encryption_config.GetEncryptionKey(false, rawPasswords)
				Expect(err).To(MatchError("encryption is disabled, but a primary encryption key was provided"))
			})
		})

		Describe("encryption is enabled", func() {
			var db *gorm.DB
			BeforeEach(func() {
				var err error
				db, err = gorm.Open("sqlite3", "test.db")
				Expect(err).NotTo(HaveOccurred())
				db_service.RunMigrations(db)
				db_service.DbConnection = db
			})

			AfterEach(func() {
				db.Close()
				os.Remove("test.db")
			})

			It("should return the primary key", func() {
				encryptionKeys := "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"80e767c6-0599-11ec-b9bf-c36874088e33\",\"label\":\"foo-foo\",\"primary\":true}]"

				key, err := encryption_config.GetEncryptionKey(true, encryptionKeys)
				Expect(err).ToNot(HaveOccurred())

				By("returning the key")
				Expect(key).ToNot(BeEmpty())
				Expect(key).ToNot(Equal("bar"))

				By("storing the encryption details")
				record, err := db_service.GetEncryptionDetailByLabel(context.Background(), "foo-foo")
				Expect(err).NotTo(HaveOccurred())
				Expect(record.Label).To(Equal("foo-foo"))
				Expect(record.Primary).To(BeTrue())
				Expect(record.Salt).NotTo(BeEmpty())
				Expect(record.Canary).NotTo(BeEmpty())
			})

			It("should use stored key when key is found in the DB", func() {
				encryptionKeys := "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"foo-foo\",\"primary\":true}]"
				key1, err := encryption_config.GetEncryptionKey(true, encryptionKeys)
				Expect(err).ToNot(HaveOccurred())

				key2, err := encryption_config.GetEncryptionKey(true, encryptionKeys)
				Expect(err).ToNot(HaveOccurred())

				By("returning the same key")
				Expect(key1).To(Equal(key2))

				By("having a single key in the DB")
				count := 0
				err = db.Model(&models.EncryptionDetail{}).Where("label = ?", "foo-foo").Count(&count).Error
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(1))
			})

			Describe("invalid encryption keys block", func() {
				It("should fail when password cannot be unmarshalled", func() {
					rawPasswords := "[{\"encryption_key\": {\"secret\":}]"

					_, err := encryption_config.GetEncryptionKey(true, rawPasswords)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("error unmarshalling encryption keys: invalid character '}' looking for beginning of value"))
				})

				It("should fail when no passwords are provided", func() {
					_, err := encryption_config.GetEncryptionKey(true, "")
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("encryption is enabled, but there was an error validating encryption keys: no encryption keys were provided"))
				})

				It("should fail when passwords are invalid", func() {
					rawPasswords := "[{\"encryption_key\": {\"secret\":\"thisisAveryLongstring\"},\"guid\":\"dae1dd13-53ed-4c90-8c11-7383b767d5c3\",\"label\":\"foo-foo\",\"primary\":false}]"

					_, err := encryption_config.GetEncryptionKey(true, rawPasswords)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("encryption is enabled, but there was an error validating encryption keys: no encryption key is marked as primary"))
				})
			})
		})
	})
})
