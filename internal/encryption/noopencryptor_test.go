package encryption_test

import (
	. "github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NoopEncryptor", func() {
	var encryptor NoopEncryptor

	BeforeEach(func() {
		encryptor = NewNoopEncryptor()
	})

	Describe("Encrypt", func() {
		It("is a noop", func() {
			const text = "my funny text to encrypt"
			Expect(encryptor.Encrypt([]byte(text))).To(Equal(text))
		})
	})

	Describe("Decrypt", func() {
		It("is a noop", func() {
			const text = "my funny text to decrypt"
			Expect(encryptor.Decrypt(text)).To(Equal([]byte(text)))
		})
	})
})
