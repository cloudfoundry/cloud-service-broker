package compoundencryptor_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models/fakes"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/compoundencryptor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CompoundEncryptor", func() {
	var (
		compoundEncryptor       compoundencryptor.Encryptor
		primaryEncryptor        *fakes.FakeEncryptor
		secondaryEncryptorAlpha *fakes.FakeEncryptor
		secondaryEncryptorBeta  *fakes.FakeEncryptor
	)

	BeforeEach(func() {
		primaryEncryptor = &fakes.FakeEncryptor{}
		secondaryEncryptorAlpha = &fakes.FakeEncryptor{}
		secondaryEncryptorBeta = &fakes.FakeEncryptor{}
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
