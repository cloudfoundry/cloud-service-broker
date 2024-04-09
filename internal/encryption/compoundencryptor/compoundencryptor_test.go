package compoundencryptor_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/encryption/compoundencryptor"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage/storagefakes"
)

var _ = Describe("CompoundEncryptor", func() {
	var (
		compoundEncryptor       storage.Encryptor
		primaryEncryptor        *storagefakes.FakeEncryptor
		secondaryEncryptorAlpha *storagefakes.FakeEncryptor
		secondaryEncryptorBeta  *storagefakes.FakeEncryptor
	)

	BeforeEach(func() {
		primaryEncryptor = &storagefakes.FakeEncryptor{}
		secondaryEncryptorAlpha = &storagefakes.FakeEncryptor{}
		secondaryEncryptorBeta = &storagefakes.FakeEncryptor{}
		compoundEncryptor = compoundencryptor.New(primaryEncryptor, secondaryEncryptorAlpha, secondaryEncryptorBeta)
	})

	It("encrypts with the encryptor", func() {
		primaryEncryptor.EncryptReturns([]byte("mopsy"), nil)

		encrypted, err := compoundEncryptor.Encrypt([]byte("flopsy"))
		Expect(err).NotTo(HaveOccurred())
		Expect(encrypted).To(Equal([]byte("mopsy")))

		Expect(primaryEncryptor.EncryptCallCount()).To(Equal(1))
		Expect(primaryEncryptor.EncryptArgsForCall(0)).To(Equal([]byte("flopsy")))
		Expect(secondaryEncryptorAlpha.EncryptCallCount()).To(BeZero())
		Expect(secondaryEncryptorBeta.EncryptCallCount()).To(BeZero())
	})

	When("encryption with the encryptor fails", func() {
		It("fails without using any of the decryptors", func() {
			primaryEncryptor.EncryptReturns(nil, errors.New("cottontail"))

			encrypted, err := compoundEncryptor.Encrypt([]byte("flopsy"))
			Expect(err).To(MatchError(errors.New("cottontail")))
			Expect(encrypted).To(Equal([]byte{}))

			Expect(primaryEncryptor.EncryptCallCount()).To(Equal(1))
			Expect(secondaryEncryptorAlpha.EncryptCallCount()).To(BeZero())
			Expect(secondaryEncryptorBeta.EncryptCallCount()).To(BeZero())
		})
	})

	When("decryption with the first decryptor fails", func() {
		It("subsequently tries the next decryptor", func() {
			primaryEncryptor.DecryptReturns(nil, errors.New("cottontail"))
			secondaryEncryptorAlpha.DecryptReturns(nil, errors.New("peter"))
			secondaryEncryptorBeta.DecryptReturns([]byte("flopsy"), nil)

			decrypted, err := compoundEncryptor.Decrypt([]byte("mopsy"))
			Expect(err).NotTo(HaveOccurred())
			Expect(decrypted).To(Equal([]byte("flopsy")))

			Expect(primaryEncryptor.DecryptCallCount()).To(Equal(0))
			Expect(secondaryEncryptorAlpha.DecryptCallCount()).To(Equal(1))
			Expect(secondaryEncryptorAlpha.DecryptArgsForCall(0)).To(Equal([]byte("mopsy")))
			Expect(secondaryEncryptorBeta.DecryptCallCount()).To(Equal(1))
			Expect(secondaryEncryptorBeta.DecryptArgsForCall(0)).To(Equal([]byte("mopsy")))
		})

		It("returns an error if all decryptors fail", func() {
			primaryEncryptor.DecryptReturns(nil, errors.New("cottontail"))
			secondaryEncryptorAlpha.DecryptReturns(nil, errors.New("peter"))
			secondaryEncryptorBeta.DecryptReturns(nil, errors.New("flopsy"))

			decrypted, err := compoundEncryptor.Decrypt([]byte("mopsy"))
			Expect(err).To(MatchError(errors.New("flopsy")))
			Expect(decrypted).To(BeEmpty())
		})
	})
})
