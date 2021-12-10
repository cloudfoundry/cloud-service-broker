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

	Describe("ServiceInstanceDetails", func() {
		Context("GCM encryptor", func() {
			BeforeEach(func() {
				key := newKey()
				encryptor = gcmencryptor.New(key)
				models.SetEncryptor(encryptor)
			})

			Describe("SetOtherDetails", func() {
				It("marshalls json content", func() {
					otherDetails := map[string]interface{}{
						"some": []interface{}{"json", "blob", "here"},
					}
					details := models.ServiceInstanceDetails{}

					err := details.SetOtherDetails(otherDetails)
					Expect(err).ToNot(HaveOccurred())

					decryptedDetails, err := encryptor.Decrypt(details.OtherDetails)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(decryptedDetails)).To(Equal(`{"some":["json","blob","here"]}`))
				})

				It("marshalls nil into json null", func() {
					details := models.ServiceInstanceDetails{}

					err := details.SetOtherDetails(nil)

					Expect(err).ToNot(HaveOccurred())
					decryptedDetails, err := encryptor.Decrypt(details.OtherDetails)
					Expect(err).ToNot(HaveOccurred())
					Expect(decryptedDetails).To(Equal([]byte("null")))
				})
			})

			Describe("GetOtherDetails", func() {
				It("decrypts and unmarshalls json content", func() {
					encryptedDetails, _ := encryptor.Encrypt([]byte(`{"some":["json","blob","here"]}`))
					serviceInstanceDetails := models.ServiceInstanceDetails{
						OtherDetails: encryptedDetails,
					}

					var actualOtherDetails map[string]interface{}
					err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

					Expect(err).ToNot(HaveOccurred())

					var arrayOfInterface []interface{}
					arrayOfInterface = append(arrayOfInterface, "json", "blob", "here")
					expectedOtherDetails := map[string]interface{}{
						"some": arrayOfInterface,
					}
					Expect(actualOtherDetails).To(Equal(expectedOtherDetails))
				})

				It("returns nil if is empty", func() {
					serviceInstanceDetails := models.ServiceInstanceDetails{}

					var actualOtherDetails map[string]interface{}
					err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

					Expect(err).ToNot(HaveOccurred())

					Expect(actualOtherDetails).To(BeNil())
				})

			})

			It("Can decrypt what it had previously encrypted", func() {
				serviceInstanceDetails := models.ServiceInstanceDetails{}
				input := map[string]interface{}{
					"some": []string{"json", "blob", "here"},
				}
				serviceInstanceDetails.SetOtherDetails(input)

				var actualOtherDetails map[string]interface{}
				err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

				Expect(err).ToNot(HaveOccurred())

				var arrayOfInterface []interface{}
				arrayOfInterface = append(arrayOfInterface, "json", "blob", "here")
				expectedOtherDetails := map[string]interface{}{
					"some": arrayOfInterface,
				}

				Expect(actualOtherDetails).To(Equal(expectedOtherDetails))
			})
		})

		Context("Noop encryptor", func() {
			BeforeEach(func() {
				encryptor = noopencryptor.New()
				models.SetEncryptor(encryptor)
			})

			Describe("SetOtherDetails", func() {
				It("marshalls json content", func() {
					otherDetails := map[string]interface{}{
						"some": []interface{}{"json", "blob", "here"},
					}
					details := models.ServiceInstanceDetails{}

					err := details.SetOtherDetails(otherDetails)

					Expect(err).ToNot(HaveOccurred())
					Expect(details.OtherDetails).To(Equal([]byte(`{"some":["json","blob","here"]}`)))
				})

				It("marshalls nil into json null", func() {
					details := models.ServiceInstanceDetails{}

					err := details.SetOtherDetails(nil)

					Expect(err).ToNot(HaveOccurred())
					Expect(details.OtherDetails).To(Equal([]byte("null")))
				})
			})

			Describe("GetOtherDetails", func() {
				It("unmarshalls json content", func() {
					serviceInstanceDetails := models.ServiceInstanceDetails{
						OtherDetails: []byte(`{"some":["json","blob","here"]}`),
					}

					var actualOtherDetails map[string]interface{}
					err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

					Expect(err).ToNot(HaveOccurred())

					var arrayOfInterface []interface{}
					arrayOfInterface = append(arrayOfInterface, "json", "blob", "here")
					expectedOtherDetails := map[string]interface{}{
						"some": arrayOfInterface,
					}
					Expect(actualOtherDetails).To(Equal(expectedOtherDetails))
				})

				It("returns nil if is empty", func() {
					serviceInstanceDetails := models.ServiceInstanceDetails{}

					var actualOtherDetails map[string]interface{}
					err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

					Expect(err).ToNot(HaveOccurred())

					Expect(actualOtherDetails).To(BeNil())
				})

			})
		})

		Describe("errors", func() {
			Describe("SetOtherDetails", func() {
				It("returns an error if it cannot marshall", func() {

					details := models.ServiceInstanceDetails{}

					err := details.SetOtherDetails(struct {
						F func()
					}{F: func() {}})

					Expect(err).To(HaveOccurred(), "Should have returned an error")
					Expect(details.OtherDetails).To(BeEmpty())
				})

				Context("When there are errors while encrypting", func() {
					BeforeEach(func() {
						fakeEncryptor := &fakes.FakeEncryptor{}
						fakeEncryptor.EncryptReturns([]byte{}, errors.New("some error"))

						encryptor = fakeEncryptor
						models.SetEncryptor(encryptor)
					})

					It("returns an error", func() {
						details := models.ServiceInstanceDetails{}
						var someDetails []byte

						err := details.SetOtherDetails(someDetails)

						Expect(err).To(MatchError("some error"))
					})
				})
			})

			Describe("GetOtherDetails", func() {
				Context("When there are errors while unmarshalling", func() {
					BeforeEach(func() {
						fakeEncryptor := &fakes.FakeEncryptor{}
						fakeEncryptor.DecryptReturns([]byte(`{"some":"badjson", "here"]}`), nil)

						encryptor = fakeEncryptor
						models.SetEncryptor(encryptor)
					})

					It("returns an error", func() {
						serviceInstanceDetails := models.ServiceInstanceDetails{
							OtherDetails: []byte("something not nil"),
						}

						var actualOtherDetails map[string]interface{}
						err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

						Expect(err).To(MatchError(ContainSubstring("invalid character")))

						Expect(actualOtherDetails).To(BeNil())
					})
				})

				Context("When there are errors while decrypting", func() {
					BeforeEach(func() {
						fakeEncryptor := &fakes.FakeEncryptor{}
						fakeEncryptor.DecryptReturns(nil, errors.New("some error"))

						encryptor = fakeEncryptor
						models.SetEncryptor(encryptor)
					})

					It("returns an error", func() {
						details := models.ServiceInstanceDetails{
							OtherDetails: []byte("something not nil"),
						}

						var actualOtherDetails map[string]interface{}
						err := details.GetOtherDetails(&actualOtherDetails)

						Expect(err).To(MatchError("some error"))
					})
				})
			})
		})
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
