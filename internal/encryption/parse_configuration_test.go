package encryption_test

import (
	"encoding/base64"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/compoundencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/gcmencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/noopencryptor"
	. "github.com/onsi/ginkgo"
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

	When("encryption is disabled and no passwords are supplied", func() {
		It("returns the no-op encryptor", func() {
			config, err := encryption.ParseConfiguration(db, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Encryptor).To(Equal(noopencryptor.New()))
			Expect(config.Changed).To(BeFalse())
			Expect(config.RotationEncryptor).To(BeNil())
			Expect(config.ConfiguredPrimaryLabel).To(Equal("none"))
			Expect(config.StoredPrimaryLabel).To(Equal("none"))
		})
	})

	When("encryption is disabled but a primary password is supplied", func() {
		const password = `[{"primary":true,"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`

		It("returns an error", func() {
			config, err := encryption.ParseConfiguration(db, false, password)
			Expect(err).To(MatchError("encryption is disabled but a primary password is set"))
			Expect(config).To(BeZero())
		})
	})

	When("encryption is disabled but was previously enabled", func() {
		const password = `[{"primary":false,"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`

		BeforeEach(func() {
			Expect(db.Create(&models.PasswordMetadata{
				Label:   "barfoo",
				Salt:    []byte("random-salt-containing-32-bytes!"),
				Canary:  "E2wsRffeAvbMceRmEE5UItxnXrakgztiTtWOJXrzk54Bpm1IwVQgxg==",
				Primary: true,
			}).Error).NotTo(HaveOccurred())
		})

		It("returns the no-op encryptor and rotation encryptor", func() {
			config, err := encryption.ParseConfiguration(db, false, password)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Encryptor).To(Equal(noopencryptor.New()))
			Expect(config.Changed).To(BeTrue())
			Expect(config.RotationEncryptor).To(BeAssignableToTypeOf(compoundencryptor.CompoundEncryptor{}))
			Expect(config.ConfiguredPrimaryLabel).To(Equal("none"))
			Expect(config.StoredPrimaryLabel).To(Equal("barfoo"))
		})
	})

	When("encryption is enabled and the primary has not changed", func() {
		const password = `[{"primary":true,"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`

		BeforeEach(func() {
			Expect(db.Create(&models.PasswordMetadata{
				Label:   "barfoo",
				Salt:    []byte("random-salt-containing-32-bytes!"),
				Canary:  "E2wsRffeAvbMceRmEE5UItxnXrakgztiTtWOJXrzk54Bpm1IwVQgxg==",
				Primary: true,
			}).Error).NotTo(HaveOccurred())
		})

		It("returns an encryptor", func() {
			config, err := encryption.ParseConfiguration(db, true, password)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))
			Expect(config.Changed).To(BeFalse())
			Expect(config.RotationEncryptor).To(BeNil())
			Expect(config.ConfiguredPrimaryLabel).To(Equal("barfoo"))
			Expect(config.StoredPrimaryLabel).To(Equal("barfoo"))

			Expect(config.Encryptor.Decrypt("cH9f37uCN/nF4wboigSqP0Xh3EkHGyJAZaCdX9kvPg==")).To(Equal([]byte("quz")))
		})
	})

	When("encryption is enabled and a primary has been supplied for the first time", func() {
		const password = `[{"primary":true,"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`

		It("returns an encryptor", func() {
			config, err := encryption.ParseConfiguration(db, true, password)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Encryptor).To(BeAssignableToTypeOf(gcmencryptor.GCMEncryptor{}))
			Expect(config.Changed).To(BeTrue())
			Expect(config.RotationEncryptor).To(BeAssignableToTypeOf(compoundencryptor.CompoundEncryptor{}))
			Expect(config.ConfiguredPrimaryLabel).To(Equal("barfoo"))
			Expect(config.StoredPrimaryLabel).To(Equal("none"))

			By("being able to encrypt with the encryptor")
			encrypted, err := config.Encryptor.Encrypt([]byte("foo"))
			Expect(err).NotTo(HaveOccurred())
			Expect(encrypted).NotTo(SatisfyAny(Equal("foo"), BeEmpty()))

			By("being able to decrypt encrypted values with the rotation encryptor")
			decrypted, err := config.RotationEncryptor.Decrypt(encrypted)
			Expect(err).NotTo(HaveOccurred())
			Expect(decrypted).To(Equal([]byte("foo")))

			By("being able to `decrypt` plaintext with the rotation encryptor")
			decrypted, err = config.RotationEncryptor.Decrypt("bar")
			Expect(err).NotTo(HaveOccurred())
			Expect(decrypted).To(Equal([]byte("bar")))
		})
	})

	When("encryption is enabled and a different primary has been supplied", func() {
		BeforeEach(func() {
			Expect(db.Create(&models.PasswordMetadata{
				Label:   "barfoo",
				Salt:    []byte("random-salt-containing-32-bytes!"),
				Canary:  "E2wsRffeAvbMceRmEE5UItxnXrakgztiTtWOJXrzk54Bpm1IwVQgxg==",
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

				By("being able to encrypt with the encryptor")
				encrypted, err := config.Encryptor.Encrypt([]byte("foo"))
				Expect(err).NotTo(HaveOccurred())
				Expect(encrypted).NotTo(SatisfyAny(Equal("foo"), BeEmpty()))

				By("being able to decrypt encrypted values with the rotation encryptor")
				decrypted, err := config.RotationEncryptor.Decrypt(encrypted)
				Expect(err).NotTo(HaveOccurred())
				Expect(decrypted).To(Equal([]byte("foo")))

				By("being able use rotation encryptor to decrypt a value encrypted with the stored primary")
				decrypted, err = config.RotationEncryptor.Decrypt("E2wsRffeAvbMceRmEE5UItxnXrakgztiTtWOJXrzk54Bpm1IwVQgxg==")
				Expect(err).NotTo(HaveOccurred())
				Expect(decrypted).To(Equal([]byte("canary value")))

				By("not being able to use the rotation encryptor to `decrypt` plaintext")
				_, err = config.RotationEncryptor.Decrypt("bar")
				Expect(err).To(MatchError(base64.CorruptInputError(0)))
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

	When("encryption is enabled but no primary password is supplied", func() {
		const password = `[{"primary":false,"label":"barfoo","password":{"secret":"averyverygoodpassword"}}]`

		It("returns an error", func() {
			config, err := encryption.ParseConfiguration(db, true, password)
			Expect(err).To(MatchError("encryption is enabled but no primary password is set"))
			Expect(config).To(BeZero())
		})
	})
})
