package encryption_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/cloudfoundry/cloud-service-broker/v3/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v3/internal/encryption"
)

var _ = Describe("DeletePasswordMetadata()", func() {
	var db *gorm.DB

	BeforeEach(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		Expect(db.Migrator().CreateTable(&models.PasswordMetadata{})).NotTo(HaveOccurred())
	})

	When("no labels are passed", func() {
		It("should not error", func() {
			Expect(encryption.DeletePasswordMetadata(db, []string{})).NotTo(HaveOccurred())
		})
	})

	When("a label list is passed", func() {
		When("all labels are in DB", func() {
			BeforeEach(func() {
				Expect(db.Create(&models.PasswordMetadata{
					Label:   "to-delete",
					Salt:    []byte("random-salt"),
					Canary:  []byte("test-value"),
					Primary: false,
				}).Error).NotTo(HaveOccurred())
				Expect(db.Create(&models.PasswordMetadata{
					Label:   "not-to-delete",
					Salt:    []byte("random-salt"),
					Canary:  []byte("test-value"),
					Primary: false,
				}).Error).NotTo(HaveOccurred())
				Expect(db.Create(&models.PasswordMetadata{
					Label:   "other-to-delete",
					Salt:    []byte("random-salt"),
					Canary:  []byte("test-value"),
					Primary: false,
				}).Error).NotTo(HaveOccurred())
			})

			It("all matching password metadata should be deleted", func() {
				err := encryption.DeletePasswordMetadata(db, []string{"to-delete", "other-to-delete"})

				Expect(err).NotTo(HaveOccurred())
				var count int64
				Expect(db.Model(&models.PasswordMetadata{}).Count(&count).Error).NotTo(HaveOccurred())
				Expect(count).To(Equal(int64(1)))
			})
		})

		When("not all labels are in DB", func() {
			BeforeEach(func() {
				Expect(db.Create(&models.PasswordMetadata{
					Label:   "to-delete",
					Salt:    []byte("random-salt"),
					Canary:  []byte("test-value"),
					Primary: false,
				}).Error).NotTo(HaveOccurred())
				Expect(db.Create(&models.PasswordMetadata{
					Label:   "not-to-delete",
					Salt:    []byte("random-salt"),
					Canary:  []byte("test-value"),
					Primary: false,
				}).Error).NotTo(HaveOccurred())
			})

			It("should delete all matching stored password metadata", func() {
				err := encryption.DeletePasswordMetadata(db, []string{"to-delete", "other-to-delete"})

				Expect(err).NotTo(HaveOccurred())
				var count int64
				Expect(db.Model(&models.PasswordMetadata{}).Count(&count).Error).NotTo(HaveOccurred())
				Expect(count).To(Equal(int64(1)))
			})
		})
	})
})
