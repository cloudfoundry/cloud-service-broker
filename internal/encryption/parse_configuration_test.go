package encryption_test

import (
	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/encryption"
	"github.com/cloudfoundry/cloud-service-broker/internal/encryption/compoundencryptor"
	"github.com/cloudfoundry/cloud-service-broker/internal/encryption/gcmencryptor"
	"github.com/cloudfoundry/cloud-service-broker/internal/encryption/noopencryptor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("ParseConfiguration()", func() {
	var db *gorm.DB

	BeforeEach(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		Expect(db.Migrator().CreateTable(&models.PasswordMetadata{})).NotTo(HaveOccurred())
	})

	When("encryption is disabled", func() {
		When("no passwords are supplied", func() {
			It("returns the no-op encryptor", func() {
				config, err := encryption.ParseConfiguration(db, false, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Encryptor).To(Equal(noopencryptor.New()))
				Expect(config.Changed).To(BeFalse())
				Expect(config.RotationEncryptor).To(BeNil())
				Expect(config.ConfiguredPrimaryLabel).To(BeEmpty())
				Expect(config.StoredPrimaryLabel).To(BeEmpty())
				Expect(config.ToDeleteLabels).To(BeZero())
			})
		})

		When("but a primary password is supplied", func() {
			const password = `[{"primary":true,"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`

			It("returns an error", func() {
				config, err := encryption.ParseConfiguration(db, false, password)
				Expect(err).To(MatchError("encryption is disabled but a primary password is set"))
				Expect(config).To(BeZero())
			})
		})

		When("but was previously enabled", func() {
			const password = `[{"primary":false,"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`

			BeforeEach(func() {
				encryptedCanary := []byte{73, 191, 136, 182, 178, 54, 18, 6, 195, 170, 245, 114, 29, 34, 193, 95, 213, 107, 30, 23, 38, 202, 37, 226, 118, 10, 247, 73, 117, 96, 27, 238, 210, 27, 46, 196, 161, 100, 254, 5}
				Expect(db.Create(&models.PasswordMetadata{
					Label:   "barfoo",
					Salt:    []byte("random-salt-containing-32-bytes!"),
					Canary:  encryptedCanary,
					Primary: true,
				}).Error).NotTo(HaveOccurred())
			})

			It("returns the no-op encryptor with a rotation encryptor", func() {
				config, err := encryption.ParseConfiguration(db, false, password)
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Encryptor).To(Equal(noopencryptor.New()))
				Expect(config.Changed).To(BeTrue())
				Expect(config.RotationEncryptor).To(BeAssignableToTypeOf(compoundencryptor.CompoundEncryptor{}))
				Expect(config.ConfiguredPrimaryLabel).To(BeEmpty())
				Expect(config.StoredPrimaryLabel).To(Equal("barfoo"))
				Expect(config.ToDeleteLabels).To(BeZero())
			})
		})

		When("a stored non-primary password is not provided", func() {
			BeforeEach(func() {
				Expect(db.Create(&models.PasswordMetadata{
					Label:   "barfoo",
					Salt:    []byte("random-salt-containing-32-bytes!"),
					Canary:  []byte("E2wsRffeAvbMceRmEE5UItxnXrakgztiTtWOJXrzk54Bpm1IwVQgxg=="),
					Primary: false,
				}).Error).NotTo(HaveOccurred())
			})

			It("is marked to delete", func() {
				config, err := encryption.ParseConfiguration(db, false, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(config.ToDeleteLabels).To(ContainElements("barfoo"))
			})
		})
	})

	When("encryption is enabled", func() {
		When("the primary has not changed", func() {
			const password = `[{"primary":true,"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`

			BeforeEach(func() {
				encryptedCanary := []byte{130, 100, 227, 172, 226, 139, 19, 69, 68, 165, 60, 67, 132, 158, 234, 45, 52, 5, 57, 243, 5, 41, 33, 170, 30, 52, 47, 204, 3, 137, 96, 132, 16, 24, 184, 33, 241, 24, 149, 35}
				Expect(db.Create(&models.PasswordMetadata{
					Label:   "barfoo",
					Salt:    []byte("random-salt-containing-32-bytes!"),
					Canary:  encryptedCanary,
					Primary: true,
				}).Error).NotTo(HaveOccurred())
			})

			It("returns an encryptor", func() {
				quzAsEncryptedBlob := []byte{59, 21, 133, 191, 122, 237, 117, 45, 137, 121, 21, 128, 28, 100, 131, 163, 91, 252, 73, 20, 74, 104, 237, 20, 103, 53, 207, 52, 154, 189, 66}
				config, err := encryption.ParseConfiguration(db, true, password)
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))
				Expect(config.Changed).To(BeFalse())
				Expect(config.RotationEncryptor).To(BeNil())
				Expect(config.ConfiguredPrimaryLabel).To(Equal("barfoo"))
				Expect(config.StoredPrimaryLabel).To(Equal("barfoo"))
				Expect(config.ToDeleteLabels).To(BeZero())
				Expect(config.Encryptor.Decrypt(quzAsEncryptedBlob)).To(Equal([]byte("quz")))
			})
		})

		When("a primary has been supplied for the first time", func() {
			const password = `[{"primary":true,"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`

			It("returns an encryptor", func() {
				config, err := encryption.ParseConfiguration(db, true, password)
				Expect(err).NotTo(HaveOccurred())
				Expect(config.Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))
				Expect(config.Changed).To(BeTrue())
				Expect(config.RotationEncryptor).To(BeAssignableToTypeOf(compoundencryptor.CompoundEncryptor{}))
				Expect(config.ConfiguredPrimaryLabel).To(Equal("barfoo"))
				Expect(config.StoredPrimaryLabel).To(BeEmpty())
				Expect(config.ToDeleteLabels).To(BeZero())

				By("being able to encrypt with the encryptor")
				encrypted, err := config.Encryptor.Encrypt([]byte("foo"))
				Expect(err).NotTo(HaveOccurred())
				Expect(encrypted).NotTo(SatisfyAny(Equal("foo"), BeEmpty()))

				By("being able to decrypt encrypted values with the rotation encryptor")
				decrypted, err := config.RotationEncryptor.Decrypt(encrypted)
				Expect(err).NotTo(HaveOccurred())
				Expect(decrypted).To(Equal([]byte("foo")))

				By("being able to `decrypt` plaintext with the rotation encryptor")
				decrypted, err = config.RotationEncryptor.Decrypt([]byte("bar"))
				Expect(err).NotTo(HaveOccurred())
				Expect(decrypted).To(Equal([]byte("bar")))
			})
		})

		When("a different primary has been supplied", func() {
			var encryptedCanary []byte

			BeforeEach(func() {
				encryptedCanary = []byte{38, 164, 195, 221, 176, 254, 121, 38, 51, 178, 60, 148, 18, 102, 146, 111, 95, 42, 233, 14, 42, 144, 50, 24, 159, 21, 162, 43, 230, 65, 211, 93, 219, 85, 11, 27, 77, 78, 210, 162}

				Expect(db.Create(&models.PasswordMetadata{
					Label:   "barfoo",
					Salt:    []byte("random-salt-containing-32-bytes!"),
					Canary:  encryptedCanary,
					Primary: true,
				}).Error).NotTo(HaveOccurred())
			})

			Context("previous primary value supplied too", func() {
				It("returns an encryptor and a rotation encryptor", func() {
					const password = `[{"primary":false,"label":"barfoo","password":{"secret":"averyverygoodpassword"}},{"label":"supernew","password":{"secret":"supercoolnewpassword"},"primary":true}]`

					config, err := encryption.ParseConfiguration(db, true, password)
					Expect(err).NotTo(HaveOccurred())
					Expect(config.Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))
					Expect(config.Changed).To(BeTrue())
					Expect(config.RotationEncryptor).To(BeAssignableToTypeOf(compoundencryptor.CompoundEncryptor{}))
					Expect(config.ConfiguredPrimaryLabel).To(Equal("supernew"))
					Expect(config.StoredPrimaryLabel).To(Equal("barfoo"))
					Expect(len(config.ToDeleteLabels)).To(BeZero())

					By("being able to encrypt with the encryptor")
					encrypted, err := config.Encryptor.Encrypt([]byte("foo"))
					Expect(err).NotTo(HaveOccurred())
					Expect(encrypted).NotTo(SatisfyAny(Equal("foo"), BeEmpty()))

					By("being able to decrypt encrypted values with the rotation encryptor")
					decrypted, err := config.RotationEncryptor.Decrypt(encrypted)
					Expect(err).NotTo(HaveOccurred())
					Expect(decrypted).To(Equal([]byte("foo")))

					By("being able use rotation encryptor to decrypt a value encrypted with the stored primary")
					decrypted, err = config.RotationEncryptor.Decrypt(encryptedCanary)
					Expect(err).NotTo(HaveOccurred())
					Expect(decrypted).To(Equal([]byte("canary value")))

					By("not being able to use the rotation encryptor to `decrypt` plaintext")
					_, err = config.RotationEncryptor.Decrypt([]byte("bar"))
					Expect(err).To(MatchError("malformed ciphertext"))
				})
			})

			Context("previous primary value not supplied", func() {
				It("returns an error", func() {
					const password = `[{"label":"supernew","password":{"secret":"supercoolnewpassword"},"primary":true}]`

					config, err := encryption.ParseConfiguration(db, true, password)
					Expect(err).To(MatchError(`the password labelled "barfoo" must be supplied to decrypt the database`))
					Expect(config).To(BeZero())
				})
			})
		})

		Context("rotation did not complete successfully", func() {
			When("the primary has not been stored as primary", func() {
				var encryptedCanary []byte

				BeforeEach(func() {
					encryptedCanary = []byte{190, 41, 181, 143, 95, 250, 158, 190, 25, 39, 45, 52, 26, 2, 67, 182, 4, 118, 144, 6, 30, 67, 150, 143, 20, 242, 15, 133, 160, 108, 38, 57, 102, 39, 119, 25, 40, 246, 75, 246}
					Expect(db.Create(&models.PasswordMetadata{
						Label:   "barfoo",
						Salt:    []byte("random-salt-containing-32-bytes!"),
						Canary:  encryptedCanary,
						Primary: false,
					}).Error).NotTo(HaveOccurred())
				})

				It("returns an encryptor and a rotation encryptor including the new primary", func() {
					const password = `[{"primary":true,"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`

					config, err := encryption.ParseConfiguration(db, true, password)
					Expect(err).NotTo(HaveOccurred())
					Expect(config.Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))
					Expect(config.Changed).To(BeTrue())
					Expect(config.RotationEncryptor).To(BeAssignableToTypeOf(compoundencryptor.CompoundEncryptor{}))
					Expect(config.ConfiguredPrimaryLabel).To(Equal("barfoo"))
					Expect(config.StoredPrimaryLabel).To(BeEmpty())
					Expect(config.ToDeleteLabels).To(BeZero())

					By("being able to encrypt with the encryptor")
					encrypted, err := config.Encryptor.Encrypt([]byte("foo"))
					Expect(err).NotTo(HaveOccurred())
					Expect(encrypted).NotTo(SatisfyAny(Equal("foo"), BeEmpty()))

					By("being able to decrypt encrypted values with the rotation encryptor")
					decrypted, err := config.RotationEncryptor.Decrypt(encrypted)
					Expect(err).NotTo(HaveOccurred())
					Expect(decrypted).To(Equal([]byte("foo")))

					By("being able use rotation encryptor to decrypt a value encrypted with the stored primary")
					decrypted, err = config.RotationEncryptor.Decrypt(encryptedCanary)
					Expect(err).NotTo(HaveOccurred())
					Expect(decrypted).To(Equal([]byte("canary value")))

					By("being able to use the rotation encryptor to `decrypt` plaintext")
					decrypted, err = config.RotationEncryptor.Decrypt([]byte("bar"))
					Expect(err).NotTo(HaveOccurred())
					Expect(decrypted).To(Equal([]byte("bar")))
				})
			})
		})

		When("no primary password is supplied", func() {
			const password = `[{"primary":false,"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`

			It("returns an error", func() {
				config, err := encryption.ParseConfiguration(db, true, password)
				Expect(err).To(MatchError("encryption is enabled but no primary password is set"))
				Expect(config).To(BeZero())
			})
		})

		When("a stored non-primary password is not provided", func() {
			BeforeEach(func() {
				Expect(db.Create(&models.PasswordMetadata{
					Label:   "to-delete",
					Salt:    []byte("random-salt-containing-32-bytes!"),
					Canary:  []byte("E2wsRffeAvbMceRmEE5UItxnXrakgztiTtWOJXrzk54Bpm1IwVQgxg=="),
					Primary: false,
				}).Error).NotTo(HaveOccurred())
			})

			It("is marked to delete", func() {
				const passwords = `[{"primary":true,"label":"barfoo","password":{"secret":"averyverygoodpassword"}}, {"primary":false,"label":"foobar","password":{"secret":"aotherveryverygoodpassword"}}]`

				config, err := encryption.ParseConfiguration(db, true, passwords)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(config.ToDeleteLabels)).To(Equal(1))
				Expect(config.ToDeleteLabels).To(ContainElements("to-delete"))
			})
		})
	})

	When("there is an error parsing the passwords", func() {
		It("returns the error", func() {
			const password = `[{"label":"firstone","password":{"secret":"tooshort"}}]`
			config, err := encryption.ParseConfiguration(db, true, password)
			Expect(err).To(MatchError("password configuration error: expected value to be 20-1024 characters long, but got length 8: [0].secret.password"))
			Expect(config).To(BeZero())
		})
	})
})
