package passwords_test

import (
	"os"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/passwords"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("Password Manager", func() {
	Describe("ProcessPasswords()", func() {
		var (
			db           *gorm.DB
			databaseFile string
		)

		BeforeEach(func() {
			fh, err := os.CreateTemp("", "")
			Expect(err).NotTo(HaveOccurred())
			databaseFile = fh.Name()
			fh.Close()

			db, err = gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
			Expect(err).NotTo(HaveOccurred())
			Expect(db_service.RunMigrations(db)).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.Remove(databaseFile)
		})

		It("returns a primary password", func() {
			const password = `[{"label":"firstone","primary":true,"password":{"secret":"averyverygoodpassword"}}]`
			passwds, err := passwords.ProcessPasswords(password, true, db)
			Expect(err).NotTo(HaveOccurred())
			Expect(passwds.Primary.Label).To(Equal("firstone"))
			Expect(passwds.Primary.Secret).To(Equal("averyverygoodpassword"))
			Expect(passwds.Secondaries).To(BeEmpty())
			Expect(passwds.ChangedPrimary).To(BeFalse())
		})

		It("returns secondary passwords", func() {
			const password = `[{"label":"barfoo","password":{"secret":"veryverysecretpassword"},"primary":false},{"label":"barbaz","password":{"secret":"anotherveryverysecretpassword"}},{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`
			passwds, err := passwords.ProcessPasswords(password, true, db)
			Expect(err).NotTo(HaveOccurred())
			Expect(passwds.Primary.Label).To(Equal("bazquz"))
			Expect(passwds.Primary.Secret).To(Equal("yetanotherveryverysecretpassword"))
			Expect(passwds.Secondaries).To(HaveLen(2))
			Expect(passwds.Secondaries[0].Label).To(Equal("barfoo"))
			Expect(passwds.Secondaries[0].Secret).To(Equal("veryverysecretpassword"))
			Expect(passwds.Secondaries[1].Label).To(Equal("barbaz"))
			Expect(passwds.Secondaries[1].Secret).To(Equal("anotherveryverysecretpassword"))
			Expect(passwds.ChangedPrimary).To(BeFalse())
		})

		It("generates different salt for each password", func() {
			const password = `[{"label":"barfoo","password":{"secret":"veryverysecretpassword"},"primary":false},{"label":"barbaz","password":{"secret":"anotherveryverysecretpassword"}},{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`
			passwds, err := passwords.ProcessPasswords(password, true, db)
			Expect(err).NotTo(HaveOccurred())
			Expect(passwds.Primary.Salt).NotTo(Equal(passwds.Secondaries[0].Salt))
			Expect(passwds.Secondaries[0].Salt).NotTo(Equal(passwds.Secondaries[1].Salt))
			Expect(passwds.Primary.Salt).NotTo(Equal(passwds.Secondaries[1].Salt))
		})

		When("no password metadata in the database", func() {
			It("stores the label, salt and canary in the database", func() {
				const password = `[{"label":"firstone","primary":true,"password":{"secret":"averyverygoodpassword"}}]`
				passwds, err := passwords.ProcessPasswords(password, true, db)
				Expect(err).NotTo(HaveOccurred())
				Expect(passwds.Primary.Label).To(Equal("firstone"))
				Expect(passwds.Primary.Secret).To(Equal("averyverygoodpassword"))
				Expect(passwds.Secondaries).To(BeEmpty())
				Expect(passwds.ChangedPrimary).To(BeFalse())

				var stored models.PasswordMetadata
				Expect(db.Where("label = ?", passwds.Primary.Label).First(&stored).Error).NotTo(HaveOccurred())
				Expect(stored.Salt).To(Equal(passwds.Primary.Salt[:]))
				Expect(stored.Canary).To(HaveLen(56))
				Expect(stored.Primary).To(BeFalse(), "should initially store as a secondary")
			})
		})

		When("password metadata in the database", func() {
			var (
				fakeSalt        string
				fakeCanary      string
				fakePrimaryFlag bool
			)

			BeforeEach(func() {
				fakeSalt = "01234567890123456789012345678901"
				fakeCanary = "5HI1YgM/2EAT/Xvd+uPKpLaQfvE7qE+2pL+XV5sOb+lTBsgchSM88w=="
				fakePrimaryFlag = true
			})

			JustBeforeEach(func() {
				db.Create(&models.PasswordMetadata{
					Label:   "firstone",
					Salt:    []byte(fakeSalt),
					Canary:  fakeCanary,
					Primary: fakePrimaryFlag,
				})
			})

			It("loads the salt from the database", func() {
				const password = `[{"label":"firstone","primary":true,"password":{"secret":"averyverygoodpassword"}}]`
				passwds, err := passwords.ProcessPasswords(password, true, db)
				Expect(err).NotTo(HaveOccurred())
				Expect(passwds.Primary.Label).To(Equal("firstone"))
				Expect(passwds.Primary.Secret).To(Equal("averyverygoodpassword"))
				Expect(passwds.Primary.Salt[:]).To(Equal([]byte(fakeSalt)))
				Expect(passwds.Secondaries).To(BeEmpty())
				Expect(passwds.ChangedPrimary).To(BeFalse())
			})

			When("the primary password has changed", func() {
				It("returns the changed flag", func() {
					const password = `[{"label":"firstone","password":{"secret":"averyverygoodpassword"},"primary":false},{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`
					passwds, err := passwords.ProcessPasswords(password, true, db)
					Expect(err).NotTo(HaveOccurred())
					Expect(passwds.Primary.Label).To(Equal("bazquz"))
					Expect(passwds.Primary.Secret).To(Equal("yetanotherveryverysecretpassword"))
					Expect(passwds.Secondaries).To(HaveLen(1))
					Expect(passwds.Secondaries[0].Label).To(Equal("firstone"))
					Expect(passwds.Secondaries[0].Secret).To(Equal("averyverygoodpassword"))

					Expect(passwds.ChangedPrimary).To(BeTrue())
				})
			})

			When("the previous primary is no longer specified", func() {
				It("returns an error", func() {
					const password = `[{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`
					_, err := passwords.ProcessPasswords(password, true, db)
					Expect(err).To(MatchError(`the previous primary password labeled "firstone" was not specified`))
				})
			})

			When("a previous secondary is no longer specified", func() {
				BeforeEach(func() {
					fakePrimaryFlag = false
				})

				It("succeeds", func() {
					const password = `[{"label":"bazquz","password":{"secret":"yetanotherveryverysecretpassword"},"primary":true}]`
					passwds, err := passwords.ProcessPasswords(password, true, db)
					Expect(err).NotTo(HaveOccurred())
					Expect(passwds.Primary.Label).To(Equal("bazquz"))
				})
			})

			When("the canary does not match", func() {
				BeforeEach(func() {
					fakeCanary = "wu1g3uwbSP0ZZjUlWfiXHNCsrK58EXIKPPsa3NdJF9K7rYaoY9khfA=="
				})

				It("returns an error", func() {
					const password = `[{"label":"firstone","primary":true,"password":{"secret":"averyverygoodpassword"}}]`
					_, err := passwords.ProcessPasswords(password, true, db)
					Expect(err).To(MatchError(`canary mismatch for password labeled "firstone" - check that the password value has not changed`))
				})
			})
		})

		When("error parsing passwords", func() {
			It("returns an error", func() {
				const password = `[{"label":"foo","password":{"secret":"01234567890123456789"},"primary":true}]`
				_, err := passwords.ProcessPasswords(password, true, db)
				Expect(err).To(MatchError(`password configuration error: expected value to be 5-20 characters long, but got length 3: [0].label`))
			})
		})

		When("no password specified", func() {
			It("fails when encryption is enabled", func() {
				const password = `[]`
				_, err := passwords.ProcessPasswords(password, true, db)
				Expect(err).To(MatchError(`encryption is enabled but no primary password is set`))
			})

			It("succeeds when encryption is disabled", func() {
				const password = `[]`
				_, err := passwords.ProcessPasswords(password, false, db)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("no primary password specified", func() {
			It("fails when encryption is enabled", func() {
				const password = `[{"label":"firstone","primary":false,"password":{"secret":"averyverygoodpassword"}}]`
				_, err := passwords.ProcessPasswords(password, true, db)
				Expect(err).To(MatchError(`encryption is enabled but no primary password is set`))
			})

			It("succeeds when encryption is disabled", func() {
				const password = `[{"label":"firstone","primary":false,"password":{"secret":"averyverygoodpassword"}}]`
				_, err := passwords.ProcessPasswords(password, false, db)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("encryption is disabled", func() {
			It("fails when a primary is specified", func() {
				const password = `[{"label":"firstone","primary":true,"password":{"secret":"averyverygoodpassword"}}]`
				_, err := passwords.ProcessPasswords(password, false, db)
				Expect(err).To(MatchError("encryption is disabled but a primary password is set"))
			})
		})
	})
})
