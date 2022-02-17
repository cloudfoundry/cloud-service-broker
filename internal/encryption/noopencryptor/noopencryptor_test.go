package noopencryptor_test

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/encryption/noopencryptor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NoopEncryptor", func() {
	var encryptor noopencryptor.NoopEncryptor

	BeforeEach(func() {
		encryptor = noopencryptor.New()
	})

	Describe("Encrypt", func() {
		It("is a noop", func() {
			const text = "my funny text to encrypt"
			Expect(encryptor.Encrypt([]byte(text))).To(Equal([]byte(text)))
		})
	})

	Describe("Decrypt", func() {
		It("is a noop", func() {
			const text = "my funny text to decrypt"
			Expect(encryptor.Decrypt([]byte(text))).To(Equal([]byte(text)))
		})
	})
})
