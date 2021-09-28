package passwordcombiner_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/gcmencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/passwordcombiner"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/passwordparser"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("Combine()", func() {
	var db *gorm.DB

	BeforeEach(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(db.Migrator().CreateTable(&models.PasswordMetadata{})).NotTo(HaveOccurred())
		Expect(err).NotTo(HaveOccurred())
	})

	Context("no stored metadata", func() {
		It("succeeds when there were no passwords", func() {
			combined, err := passwordcombiner.Combine(db, []passwordparser.PasswordEntry{}, []models.PasswordMetadata{})
			Expect(err).NotTo(HaveOccurred())
			Expect(combined).To(BeEmpty())
		})

		It("can return a password", func() {
			password := []passwordparser.PasswordEntry{
				{
					Label:  "firstone",
					Secret: "averyverygoodpassword",
				},
			}
			combined, err := passwordcombiner.Combine(db, password, []models.PasswordMetadata{})
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
			passwords := []passwordparser.PasswordEntry{
				{
					Label:  "barfoo",
					Secret: "veryverysecretpassword",
				},
				{
					Label:  "barbaz",
					Secret: "anotherveryverysecretpassword",
				},
				{
					Label:   "bazquz",
					Secret:  "yetanotherveryverysecretpassword",
					Primary: true,
				},
			}
			combined, err := passwordcombiner.Combine(db, passwords, []models.PasswordMetadata{})
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
	})

	Context("stored metadata", func() {
		var barfooSalt, barbazSalt []byte
		var storedMetadata []models.PasswordMetadata

		BeforeEach(func() {
			barfooSalt = []byte("random-salt-containing-32-bytes!")
			barbazSalt = []byte("another-random-salt-with-32bytes")
			encryptedCanary := []byte{250, 65, 162, 134, 203, 81, 170, 159, 176, 113, 29, 249, 223, 77, 187, 139, 97, 254, 110, 99, 177, 102, 234, 51, 47, 85, 126, 205, 110, 173, 159, 209, 234, 138, 66, 113, 117, 191, 211, 184}
			storedMetadata = []models.PasswordMetadata{
				{
					Label:   "barfoo",
					Salt:    barfooSalt,
					Canary:  encryptedCanary,
					Primary: false,
				},
			}
		})

		It("succeeds when there were no passwords", func() {

			combined, err := passwordcombiner.Combine(db, []passwordparser.PasswordEntry{}, storedMetadata)
			Expect(err).NotTo(HaveOccurred())
			Expect(combined).To(BeEmpty())
		})

		It("can return a password with stored salt value", func() {
			password := []passwordparser.PasswordEntry{
				{
					Label:  "barfoo",
					Secret: "averyverygoodpassword",
				},
			}

			combined, err := passwordcombiner.Combine(db, password, storedMetadata)
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
			encryptedCanary := []byte{74, 35, 85, 82, 16, 202, 239, 216, 209, 30, 158, 65, 28, 0, 77, 203, 96, 155, 20, 61, 16, 204, 81, 147, 22, 42, 144, 193, 95, 50, 47, 207, 156, 106, 219, 159, 90, 8, 13, 59}
			storedMetadata = append(storedMetadata, models.PasswordMetadata{
				Label:   "barbaz",
				Salt:    barbazSalt,
				Canary:  encryptedCanary,
				Primary: true,
			})
			passwords := []passwordparser.PasswordEntry{
				{
					Label:   "barfoo",
					Secret:  "averyverygoodpassword",
					Primary: true,
				},
				{
					Label:  "barbaz",
					Secret: "anotherveryverysecretpassword",
				},
				{
					Label:  "bazquz",
					Secret: "yetanotherveryverysecretpassword",
				},
			}

			combined, err := passwordcombiner.Combine(db, passwords, storedMetadata)
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
				password := []passwordparser.PasswordEntry{
					{
						Label:  "barfoo",
						Secret: "notthesameaslasttime",
					},
				}
				combined, err := passwordcombiner.Combine(db, password, storedMetadata)
				Expect(err).To(MatchError(`canary mismatch for password labeled "barfoo" - check that the password value has not changed`))
				Expect(combined).To(BeEmpty())
			})
		})

		When("more than one primary is stored", func() {
			It("returns an error", func() {
				storedMetadata := []models.PasswordMetadata{
					{
						Label:   "barbaz",
						Salt:    barbazSalt,
						Canary:  []byte("XVVB0psiTW1J9R/r8Sh32aY2oddKujDnNHMAzcMowrdnO+ngixJn8g=="),
						Primary: true,
					},
					{
						Label:   "anotherone",
						Salt:    barfooSalt,
						Canary:  []byte("E2wsRffeAvbMceRmEE5UItxnXrakgztiTtWOJXrzk54Bpm1IwVQgxg=="),
						Primary: true,
					},
				}

				password := []passwordparser.PasswordEntry{
					{
						Label:  "barfoo",
						Secret: "notthesameaslasttime",
					},
				}
				combined, err := passwordcombiner.Combine(db, password, storedMetadata)
				Expect(err).To(MatchError(`corrupt database - more than one primary found in table password_metadata`))
				Expect(combined).To(BeEmpty())
			})
		})

		When("a primary is stored but password not supplied", func() {
			It("returns an error", func() {
				storedMetadata = []models.PasswordMetadata{
					{
						Label:   "barbaz",
						Salt:    barbazSalt,
						Canary:  []byte("XVVB0psiTW1J9R/r8Sh32aY2oddKujDnNHMAzcMowrdnO+ngixJn8g=="),
						Primary: true,
					},
				}
				password := []passwordparser.PasswordEntry{
					{
						Label:  "barfoo",
						Secret: "averyverygoodpassword",
					},
				}

				combined, err := passwordcombiner.Combine(db, password, storedMetadata)
				Expect(err).To(MatchError(`the password labelled "barbaz" must be supplied to decrypt the database`))
				Expect(combined).To(BeEmpty())
			})
		})
	})
})
