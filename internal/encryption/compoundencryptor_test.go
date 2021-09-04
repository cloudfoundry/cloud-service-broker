package encryption_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models/fakes"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CompoundEncryptor", func() {
	var (
		compoundEncryptor       encryption.Encryptor
		primaryEncryptor        *fakes.FakeEncryptor
		secondaryEncryptorAlpha *fakes.FakeEncryptor
		secondaryEncryptorBeta  *fakes.FakeEncryptor
	)

	BeforeEach(func() {
		primaryEncryptor = &fakes.FakeEncryptor{}
		secondaryEncryptorAlpha = &fakes.FakeEncryptor{}
		secondaryEncryptorBeta = &fakes.FakeEncryptor{}
		compoundEncryptor = encryption.NewCompoundEncryptor(primaryEncryptor, secondaryEncryptorAlpha, secondaryEncryptorBeta)
	})

	It("encrypts with the primary encryptor", func() {
		primaryEncryptor.EncryptReturns("mopsy", nil)

		encrypted, err := compoundEncryptor.Encrypt([]byte("flopsy"))
		Expect(err).NotTo(HaveOccurred())
		Expect(encrypted).To(Equal("mopsy"))

		Expect(primaryEncryptor.EncryptCallCount()).To(Equal(1))
		Expect(primaryEncryptor.EncryptArgsForCall(0)).To(Equal([]byte("flopsy")))
		Expect(secondaryEncryptorAlpha.EncryptCallCount()).To(BeZero())
		Expect(secondaryEncryptorBeta.EncryptCallCount()).To(BeZero())
	})

	When("encryption with the primary fails", func() {
		It("fails without using any of the secondaries", func() {
			primaryEncryptor.EncryptReturns("", errors.New("cottontail"))

			encrypted, err := compoundEncryptor.Encrypt([]byte("flopsy"))
			Expect(err).To(MatchError(errors.New("cottontail")))
			Expect(encrypted).To(Equal(""))

			Expect(primaryEncryptor.EncryptCallCount()).To(Equal(1))
			Expect(secondaryEncryptorAlpha.EncryptCallCount()).To(BeZero())
			Expect(secondaryEncryptorBeta.EncryptCallCount()).To(BeZero())
		})
	})

	It("decrypts with the primary decryptor", func() {
		primaryEncryptor.DecryptReturns([]byte("peter"), nil)

		decrypted, err := compoundEncryptor.Decrypt("flopsy")
		Expect(err).NotTo(HaveOccurred())
		Expect(decrypted).To(Equal([]byte("peter")))

		Expect(primaryEncryptor.DecryptCallCount()).To(Equal(1))
		Expect(primaryEncryptor.DecryptArgsForCall(0)).To(Equal("flopsy"))
		Expect(secondaryEncryptorAlpha.DecryptCallCount()).To(BeZero())
		Expect(secondaryEncryptorBeta.DecryptCallCount()).To(BeZero())
	})

	When("decryption with the primary fails", func() {
		It("tries the first secondary", func() {
			primaryEncryptor.DecryptReturns(nil, errors.New("cottontail"))
			secondaryEncryptorAlpha.DecryptReturns([]byte("flopsy"), nil)

			decrypted, err := compoundEncryptor.Decrypt("peter")
			Expect(err).NotTo(HaveOccurred())
			Expect(decrypted).To(Equal([]byte("flopsy")))

			Expect(primaryEncryptor.DecryptCallCount()).To(Equal(1))
			Expect(primaryEncryptor.DecryptArgsForCall(0)).To(Equal("peter"))
			Expect(secondaryEncryptorAlpha.DecryptCallCount()).To(Equal(1))
			Expect(secondaryEncryptorAlpha.DecryptArgsForCall(0)).To(Equal("peter"))
			Expect(secondaryEncryptorBeta.DecryptCallCount()).To(BeZero())
		})

		It("subsequently tries the next secondary", func() {
			primaryEncryptor.DecryptReturns(nil, errors.New("cottontail"))
			secondaryEncryptorAlpha.DecryptReturns(nil, errors.New("peter"))
			secondaryEncryptorBeta.DecryptReturns([]byte("flopsy"), nil)

			decrypted, err := compoundEncryptor.Decrypt("mopsy")
			Expect(err).NotTo(HaveOccurred())
			Expect(decrypted).To(Equal([]byte("flopsy")))

			Expect(primaryEncryptor.DecryptCallCount()).To(Equal(1))
			Expect(primaryEncryptor.DecryptArgsForCall(0)).To(Equal("mopsy"))
			Expect(secondaryEncryptorAlpha.DecryptCallCount()).To(Equal(1))
			Expect(secondaryEncryptorAlpha.DecryptArgsForCall(0)).To(Equal("mopsy"))
			Expect(secondaryEncryptorBeta.DecryptCallCount()).To(Equal(1))
			Expect(secondaryEncryptorBeta.DecryptArgsForCall(0)).To(Equal("mopsy"))
		})

		It("returns an error if all secondaries fail", func() {
			primaryEncryptor.DecryptReturns(nil, errors.New("cottontail"))
			secondaryEncryptorAlpha.DecryptReturns(nil, errors.New("peter"))
			secondaryEncryptorBeta.DecryptReturns(nil, errors.New("flopsy"))

			decrypted, err := compoundEncryptor.Decrypt("mopsy")
			Expect(err).To(MatchError(errors.New("flopsy")))
			Expect(decrypted).To(BeEmpty())
		})
	})
})
