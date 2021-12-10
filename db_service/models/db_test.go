package models_test

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models/fakes"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/gcmencryptor"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption/noopencryptor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func newKey() [32]byte {
	dbKey := make([]byte, 32)
	io.ReadFull(rand.Reader, dbKey)
	return sha256.Sum256(dbKey)
}

var _ = Describe("Db", func() {
	var encryptor models.Encryptor

	AfterEach(func() {
		models.SetEncryptor(nil)
	})

	Describe("TerraformDeployment", func() {
		const plaintext = "plaintext"

		Context("GCM encryptor", func() {
			BeforeEach(func() {
				key := newKey()
				encryptor = gcmencryptor.New(key)
				models.SetEncryptor(encryptor)
			})

			Describe("SetWorkspace", func() {
				It("encrypts the workspace", func() {
					By("making sure it's no longer in plaintext")
					var t models.TerraformDeployment
					err := t.SetWorkspace(plaintext)
					Expect(err).NotTo(HaveOccurred())
					Expect(t.Workspace).NotTo(Equal(plaintext))

					By("being able to decrypt it")
					p, err := encryptor.Decrypt(t.Workspace)
					Expect(err).NotTo(HaveOccurred())
					Expect(p).To(Equal([]byte(plaintext)))
				})
			})

			Describe("GetWorkspace", func() {
				var t models.TerraformDeployment

				BeforeEach(func() {
					err := t.SetWorkspace(plaintext)
					Expect(err).NotTo(HaveOccurred())
				})

				It("can read a previously encrypted workspace", func() {
					By("checking that it's not in plaintext")
					Expect(t.Workspace).NotTo(Equal(plaintext))

					By("reading the plaintext")
					p, err := t.GetWorkspace()
					Expect(err).NotTo(HaveOccurred())
					Expect(p).To(Equal(plaintext))
				})
			})
		})

		Context("Noop encryptor", func() {
			BeforeEach(func() {
				encryptor = noopencryptor.New()
				models.SetEncryptor(encryptor)
			})

			Describe("SetWorkspace", func() {
				It("sets the workspace in plaintext", func() {
					var t models.TerraformDeployment
					err := t.SetWorkspace(plaintext)
					Expect(err).NotTo(HaveOccurred())
					Expect(t.Workspace).To(Equal([]byte(plaintext)))
				})
			})

			Describe("GetWorkspace", func() {
				It("reads the plaintext workspace", func() {
					t := models.TerraformDeployment{Workspace: []byte(plaintext)}
					v, err := t.GetWorkspace()
					Expect(err).NotTo(HaveOccurred())
					Expect(v).To(Equal(plaintext))
				})
			})
		})

		Describe("errors", func() {
			Describe("SetWorkspace", func() {
				BeforeEach(func() {
					fakeEncryptor := &fakes.FakeEncryptor{}
					fakeEncryptor.EncryptReturns([]byte{}, errors.New("fake encryption error"))

					encryptor = fakeEncryptor
					models.SetEncryptor(encryptor)
				})

				It("returns encryption errors", func() {
					var t models.TerraformDeployment
					err := t.SetWorkspace(plaintext)
					Expect(err).To(MatchError("fake encryption error"))
				})
			})

			Describe("GetWorkspace", func() {
				BeforeEach(func() {
					fakeEncryptor := &fakes.FakeEncryptor{}
					fakeEncryptor.DecryptReturns(nil, errors.New("fake decryption error"))

					encryptor = fakeEncryptor
					models.SetEncryptor(encryptor)
				})

				It("returns decryption errors", func() {
					t := models.TerraformDeployment{Workspace: []byte(plaintext)}
					v, err := t.GetWorkspace()
					Expect(err).To(MatchError("fake decryption error"))
					Expect(v).To(BeEmpty())
				})
			})
		})
	})
})
