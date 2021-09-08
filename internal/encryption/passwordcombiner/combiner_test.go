package passwordcombiner_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/gcmencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/passwordcombiner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("CombineWithStoredMetadata()", func() {
	var db *gorm.DB

	BeforeEach(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(db.Migrator().CreateTable(&models.PasswordMetadata{})).NotTo(HaveOccurred())
		Expect(err).NotTo(HaveOccurred())
	})

	Context("no stored metadata", func() {
		It("succeeds when there were no passwords", func() {
			combined, err := passwordcombiner.CombineWithStoredMetadata(db, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(combined).To(BeEmpty())
		})

		It("can return a password", func() {
			const password = `[{"label":"firstone","password":{"secret":"averyverygoodpassword"}}]`
			combined, err := passwordcombiner.CombineWithStoredMetadata(db, password)
			Expect(err).NotTo(HaveOccurred())

			By("returning the password")
			Expect(combined).To(HaveLen(1))
			Expect(combined[0].Label).To(Equal("firstone"))
			Expect(combined[0].Secret).To(Equal("averyverygoodpassword"))
			Expect(combined[0].Salt).To(HaveLen(32))
			Expect(combined[0].Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))

			By("identifying no primary")
			primary, ok := combined.ConfiguredPrimary()
			Expect(ok).To(BeFalse())
			Expect(primary).To(BeZero())

			By("storing the password metadata in the database")
			var stored models.PasswordMetadata
			Expect(db.First(&stored).Error).NotTo(HaveOccurred())
			Expect(stored.Label).To(Equal("firstone"))
			Expect(stored.Salt).To(Equal(combined[0].Salt))
			Expect(stored.Primary).To(BeFalse())
			Expect(stored.Canary).NotTo(BeEmpty())
		})

		It("can return multiple passwords", func() {
			const passwords = `[{"label":"barfoo","password":{"secret":"veryverysecretpassword"},"primary":false},{"label":"barbaz","password":{"secret":"anotherveryverysecretpassword"}},{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`
			combined, err := passwordcombiner.CombineWithStoredMetadata(db, passwords)
			Expect(err).NotTo(HaveOccurred())

			By("returning the passwords")
			Expect(combined).To(HaveLen(3))
			Expect(combined[0].Label).To(Equal("barfoo"))
			Expect(combined[0].Secret).To(Equal("veryverysecretpassword"))
			Expect(combined[0].Salt).To(HaveLen(32))
			Expect(combined[0].Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))
			Expect(combined[1].Label).To(Equal("barbaz"))
			Expect(combined[1].Secret).To(Equal("anotherveryverysecretpassword"))
			Expect(combined[1].Salt).To(HaveLen(32))
			Expect(combined[1].Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))
			Expect(combined[2].Label).To(Equal("bazquz"))
			Expect(combined[2].Secret).To(Equal("yetanotherveryverysecretpassword"))
			Expect(combined[2].Salt).To(HaveLen(32))
			Expect(combined[2].Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))

			By("assigning unique salt to each password")
			Expect(combined[0].Salt).NotTo(Equal(combined[2].Salt))
			Expect(combined[1].Salt).NotTo(Equal(combined[2].Salt))
			Expect(combined[1].Salt).NotTo(Equal(combined[2].Salt))

			By("identifying the primary")
			primary, ok := combined.ConfiguredPrimary()
			Expect(ok).To(BeTrue())
			Expect(primary.Label).To(Equal("bazquz"))
			Expect(primary.Secret).To(Equal("yetanotherveryverysecretpassword"))
			Expect(primary.Salt).To(HaveLen(32))

			By("storing the password metadata in the database")
			var stored []models.PasswordMetadata
			Expect(db.Find(&stored).Error).NotTo(HaveOccurred())
			Expect(stored).To(HaveLen(3))
			Expect(stored[0].Label).To(Equal("barfoo"))
			Expect(stored[0].Canary).NotTo(BeEmpty())
			Expect(stored[0].Salt).To(HaveLen(32))
			Expect(stored[0].Primary).To(BeFalse())
			Expect(stored[1].Label).To(Equal("barbaz"))
			Expect(stored[1].Canary).NotTo(BeEmpty())
			Expect(stored[1].Salt).To(HaveLen(32))
			Expect(stored[1].Primary).To(BeFalse())
			Expect(stored[2].Label).To(Equal("bazquz"))
			Expect(stored[2].Canary).NotTo(BeEmpty())
			Expect(stored[2].Salt).To(HaveLen(32))
			Expect(stored[2].Primary).To(BeFalse())
		})

		When("there is an error parsing the passwords", func() {
			It("returns the error", func() {
				const password = `[{"label":"firstone","password":{"secret":"tooshort"}}]`
				combined, err := passwordcombiner.CombineWithStoredMetadata(db, password)
				Expect(err).To(MatchError("password configuration error: expected value to be 20-1024 characters long, but got length 8: [0].secret.password"))
				Expect(combined).To(BeEmpty())
			})
		})
	})

	Context("stored metadata", func() {
		var barfooSalt, barbazSalt []byte

		BeforeEach(func() {
			barfooSalt = []byte("random-salt-containing-32-bytes!")
			Expect(db.Create(&models.PasswordMetadata{
				Label:   "barfoo",
				Salt:    barfooSalt,
				Canary:  "E2wsRffeAvbMceRmEE5UItxnXrakgztiTtWOJXrzk54Bpm1IwVQgxg==",
				Primary: false,
			}).Error).NotTo(HaveOccurred())

			barbazSalt = []byte("another-random-salt-with-32bytes")
		})

		It("succeeds when there were no passwords", func() {
			combined, err := passwordcombiner.CombineWithStoredMetadata(db, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(combined).To(BeEmpty())
		})

		It("can return a password with stored salt value", func() {
			const password = `[{"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`
			combined, err := passwordcombiner.CombineWithStoredMetadata(db, password)
			Expect(err).NotTo(HaveOccurred())

			By("returning the password")
			Expect(combined).To(HaveLen(1))
			Expect(combined[0].Label).To(Equal("barfoo"))
			Expect(combined[0].Secret).To(Equal("averyverygoodpassword"))
			Expect(combined[0].Salt).To(Equal(barfooSalt))
			Expect(combined[0].Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))

			By("identifying no stored primary")
			primary, ok := combined.StoredPrimary()
			Expect(ok).To(BeFalse())
			Expect(primary).To(BeZero())
		})

		It("can return multiple passwords with stored salt values", func() {
			Expect(db.Create(&models.PasswordMetadata{
				Label:   "barbaz",
				Salt:    barbazSalt,
				Canary:  "XVVB0psiTW1J9R/r8Sh32aY2oddKujDnNHMAzcMowrdnO+ngixJn8g==",
				Primary: true,
			}).Error).NotTo(HaveOccurred())

			const passwords = `[{"label":"barfoo","password":{"secret":"averyverygoodpassword"},"primary":true},{"label":"barbaz","password":{"secret":"anotherveryverysecretpassword"}},{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"}}]`
			combined, err := passwordcombiner.CombineWithStoredMetadata(db, passwords)
			Expect(err).NotTo(HaveOccurred())

			By("returning the passwords")
			Expect(combined).To(HaveLen(3))
			Expect(combined[0].Label).To(Equal("barfoo"))
			Expect(combined[0].Secret).To(Equal("averyverygoodpassword"))
			Expect(combined[0].Salt).To(Equal(barfooSalt))
			Expect(combined[0].Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))
			Expect(combined[1].Label).To(Equal("barbaz"))
			Expect(combined[1].Secret).To(Equal("anotherveryverysecretpassword"))
			Expect(combined[1].Salt).To(Equal(barbazSalt))
			Expect(combined[1].Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))
			Expect(combined[2].Label).To(Equal("bazquz"))
			Expect(combined[2].Secret).To(Equal("yetanotherveryverysecretpassword"))
			Expect(combined[2].Salt).To(HaveLen(32))
			Expect(combined[2].Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))

			By("identifying the primary amongst the stored metadata")
			primary, ok := combined.StoredPrimary()
			Expect(ok).To(BeTrue())
			Expect(primary.Label).To(Equal("barbaz"))
			Expect(primary.Secret).To(Equal("anotherveryverysecretpassword"))
			Expect(primary.Salt).To(Equal(barbazSalt))

			By("identifying the primary amongst the supplied passwords")
			primary, ok = combined.ConfiguredPrimary()
			Expect(ok).To(BeTrue())
			Expect(primary.Label).To(Equal("barfoo"))
			Expect(primary.Secret).To(Equal("averyverygoodpassword"))
			Expect(primary.Salt).To(Equal(barfooSalt))

			By("storing the new password metadata in the database")
			var stored models.PasswordMetadata
			Expect(db.Where("label = ?", "bazquz").First(&stored).Error).NotTo(HaveOccurred())
			Expect(stored.Canary).NotTo(BeEmpty())
			Expect(stored.Salt).To(HaveLen(32))
			Expect(stored.Primary).To(BeFalse())
		})

		When("password changed value", func() {
			It("returns an error", func() {
				const password = `[{"label":"barfoo","password":{"secret":"notthesameaslasttime"}}]`
				combined, err := passwordcombiner.CombineWithStoredMetadata(db, password)
				Expect(err).To(MatchError(`canary mismatch for password labeled "barfoo" - check that the password value has not changed`))
				Expect(combined).To(BeEmpty())
			})
		})

		When("more than one primary is stored", func() {
			BeforeEach(func() {
				Expect(db.Create(&models.PasswordMetadata{
					Label:   "barbaz",
					Salt:    barbazSalt,
					Canary:  "XVVB0psiTW1J9R/r8Sh32aY2oddKujDnNHMAzcMowrdnO+ngixJn8g==",
					Primary: true,
				}).Error).NotTo(HaveOccurred())

				Expect(db.Create(&models.PasswordMetadata{
					Label:   "anotherone",
					Salt:    barfooSalt,
					Canary:  "E2wsRffeAvbMceRmEE5UItxnXrakgztiTtWOJXrzk54Bpm1IwVQgxg==",
					Primary: true,
				}).Error).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				const password = `[{"label":"barfoo","password":{"secret":"notthesameaslasttime"}}]`
				combined, err := passwordcombiner.CombineWithStoredMetadata(db, password)
				Expect(err).To(MatchError(`corrupt database - more than one primary found in table password_metadata`))
				Expect(combined).To(BeEmpty())
			})
		})

		When("a primary is stored but password not supplied", func() {
			BeforeEach(func() {
				Expect(db.Create(&models.PasswordMetadata{
					Label:   "barbaz",
					Salt:    barbazSalt,
					Canary:  "XVVB0psiTW1J9R/r8Sh32aY2oddKujDnNHMAzcMowrdnO+ngixJn8g==",
					Primary: true,
				}).Error).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				const password = `[{"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`
				combined, err := passwordcombiner.CombineWithStoredMetadata(db, password)
				Expect(err).To(MatchError(`the password labelled "barbaz" must be supplied to decrypt the database`))
				Expect(combined).To(BeEmpty())
			})
		})
	})
})
