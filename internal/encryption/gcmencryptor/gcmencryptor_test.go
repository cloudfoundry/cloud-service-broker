package gcmencryptor_test

import (
	"crypto/rand"
	"crypto/sha256"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/encryption/gcmencryptor"
)

var _ = Describe("GCMEncryptor", func() {
	var encryptor gcmencryptor.GCMEncryptor
	BeforeEach(func() {
		key := newKey()
		encryptor = gcmencryptor.New(key)
	})

	It("can decrypt what it encrypted", func() {
		encrypted, err := encryptor.Encrypt([]byte("Text to Encrypt"))
		Expect(err).ToNot(HaveOccurred())
		Expect(encrypted).ToNot(ContainSubstring("Encrypt"))

		decrypted, err := encryptor.Decrypt(encrypted)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(decrypted)).To(Equal("Text to Encrypt"))

	})

	Describe("Encrypt", func() {
		It("encrypts data using a nonce", func() {
			By("encrypting a few times and checking we dont get the same result")
			const textToEncrypt = "Text to Encrypt"
			result1, err := encryptor.Encrypt([]byte(textToEncrypt))
			Expect(err).ToNot(HaveOccurred())
			result2, err := encryptor.Encrypt([]byte(textToEncrypt))
			Expect(err).ToNot(HaveOccurred())
			Expect(result2).ToNot(Equal(result1))
			result3, err := encryptor.Encrypt([]byte(textToEncrypt))
			Expect(err).ToNot(HaveOccurred())
			Expect(result3).ToNot(Equal(result1))
			Expect(result3).ToNot(Equal(result2))
		})

		It("panics when run on an uninitialised encryptor", func() {
			Expect(func() { _, _ = gcmencryptor.GCMEncryptor{}.Encrypt([]byte("foo")) }).To(Panic())
		})
	})

	Describe("Decrypt", func() {

		It("fails if text is malformed", func() {
			result, err := encryptor.Decrypt([]byte("shorter"))
			Expect(err).To(MatchError("malformed ciphertext"))
			Expect(result).To(BeNil())
		})

		It("fails if text is corrupted", func() {
			result, err := encryptor.Decrypt([]byte("longtextthatdoesnotcontainthetag"))
			Expect(err).To(MatchError("cipher: message authentication failed"))
			Expect(result).To(BeNil())
		})

		It("panics when run on an uninitialised encryptor", func() {
			Expect(func() { _, _ = gcmencryptor.GCMEncryptor{}.Decrypt([]byte("foo")) }).To(Panic())
		})
	})
})

func newKey() [32]byte {
	dbKey := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, dbKey)
	return sha256.Sum256(dbKey)
}
