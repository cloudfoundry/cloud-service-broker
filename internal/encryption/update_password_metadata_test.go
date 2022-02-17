package encryption_test

import (
	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/encryption"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseConfiguration()", func() {
	var db *gorm.DB

	BeforeEach(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		Expect(db.Migrator().CreateTable(&models.PasswordMetadata{})).NotTo(HaveOccurred())
	})

	When("new primary password is different", func() {
		BeforeEach(func() {
			Expect(db.Create(&models.PasswordMetadata{
				Label:   "previous-primary",
				Salt:    []byte("random-salt"),
				Canary:  []byte("test-value"),
				Primary: true,
			}).Error).NotTo(HaveOccurred())
			Expect(db.Create(&models.PasswordMetadata{
				Label:   "new-primary",
				Salt:    []byte("other-random-salt"),
				Canary:  []byte("other-test-value"),
				Primary: false,
			}).Error).NotTo(HaveOccurred())
		})

		It("swaps the primary flag", func() {
			err := encryption.UpdatePasswordMetadata(db, "new-primary")
			Expect(err).NotTo(HaveOccurred())

			var oldPrimary models.PasswordMetadata
			Expect(db.Where("label = ?", "previous-primary").First(&oldPrimary).Error).NotTo(HaveOccurred())
			Expect(oldPrimary.Primary).To(BeFalse())

			var newPrimary models.PasswordMetadata
			Expect(db.Where("label = ?", "new-primary").First(&newPrimary).Error).NotTo(HaveOccurred())
			Expect(newPrimary.Primary).To(BeTrue())
		})
	})

	When("new primary password is the same as existing password", func() {
		BeforeEach(func() {
			Expect(db.Create(&models.PasswordMetadata{
				Label:   "primary",
				Salt:    []byte("random-salt"),
				Canary:  []byte("test-value"),
				Primary: true,
			}).Error).NotTo(HaveOccurred())
		})

		It("does not change the primary flag", func() {
			var beforePrimary models.PasswordMetadata
			Expect(db.Where("label = ?", "primary").First(&beforePrimary).Error).NotTo(HaveOccurred())
			updateTime := beforePrimary.UpdatedAt

			err := encryption.UpdatePasswordMetadata(db, "primary")
			Expect(err).NotTo(HaveOccurred())

			var primary models.PasswordMetadata
			Expect(db.Where("label = ?", "primary").First(&primary).Error).NotTo(HaveOccurred())
			Expect(primary.Primary).To(BeTrue())
			Expect(primary.UpdatedAt).To(Equal(updateTime))
		})
	})

	When("there is no new primary password", func() {
		BeforeEach(func() {
			Expect(db.Create(&models.PasswordMetadata{
				Label:   "primary",
				Salt:    []byte("random-salt"),
				Canary:  []byte("test-value"),
				Primary: true,
			}).Error).NotTo(HaveOccurred())
		})

		When("new primary label is empty", func() {
			It("sets all existing primaries flags to false", func() {
				err := encryption.UpdatePasswordMetadata(db, "")
				Expect(err).NotTo(HaveOccurred())

				var primary models.PasswordMetadata
				Expect(db.Where("label = ?", "primary").First(&primary).Error).NotTo(HaveOccurred())
				Expect(primary.Primary).To(BeFalse())
			})
		})
	})

	When("new primary password cannot be found", func() {
		It("returns an error", func() {
			err := encryption.UpdatePasswordMetadata(db, "some-pass")
			Expect(err).To(MatchError("cannot find metadata for password labelled \"some-pass\""))
		})
	})
})
