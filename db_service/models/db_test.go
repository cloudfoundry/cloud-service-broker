package models_test

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
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

	Describe("ServiceBindingCredentials", func() {
		Context("GCM encryptor", func() {
			BeforeEach(func() {
				key := newKey()
				encryptor = gcmencryptor.New(key)
				models.SetEncryptor(encryptor)
			})

			Describe("SetOtherDetails", func() {
				It("encrypts the field", func() {
					const expectedJSON = `{"some":["json","blob","here"]}`
					otherDetails := map[string]interface{}{
						"some": []interface{}{"json", "blob", "here"},
					}

					credentials := models.ServiceBindingCredentials{}
					err := credentials.SetOtherDetails(otherDetails)
					Expect(err).ToNot(HaveOccurred())

					By("checking that it's no longer in plaintext")
					Expect(credentials.OtherDetails).ToNot(Equal(expectedJSON))

					By("being able to decrypt to get the value")
					decrypted, err := encryptor.Decrypt(credentials.OtherDetails)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(decrypted)).To(Equal(expectedJSON))
				})
			})

			Describe("GetOtherDetails", func() {
				It("can decrypt what it had previously encrypted", func() {
					By("encrypting a field")
					otherDetails := map[string]interface{}{
						"some": []interface{}{"json", "blob", "here"},
					}
					credentials := models.ServiceBindingCredentials{}
					credentials.SetOtherDetails(otherDetails)

					By("checking that it's encrypted")
					Expect(credentials.OtherDetails).ToNot(ContainSubstring("some"))

					By("decrypting that field")
					var actualOtherDetails map[string]interface{}
					err := credentials.GetOtherDetails(&actualOtherDetails)
					Expect(err).ToNot(HaveOccurred())

					Expect(actualOtherDetails).To(HaveLen(1))
					Expect(actualOtherDetails).To(HaveKeyWithValue("some", ConsistOf("json", "blob", "here")))
				})
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

					credentials := models.ServiceBindingCredentials{}
					err := credentials.SetOtherDetails(otherDetails)
					Expect(err).ToNot(HaveOccurred())

					Expect(credentials.OtherDetails).To(Equal([]byte(`{"some":["json","blob","here"]}`)))
				})

				It("marshalls nil into json null", func() {
					credentials := models.ServiceBindingCredentials{}
					err := credentials.SetOtherDetails(nil)
					Expect(err).ToNot(HaveOccurred())

					Expect(credentials.OtherDetails).To(Equal([]byte("null")))
				})

				It("returns an error if it cannot marshall", func() {
					credentials := models.ServiceBindingCredentials{}
					err := credentials.SetOtherDetails(struct {
						F func()
					}{F: func() {}})

					Expect(err).To(MatchError(ContainSubstring("unsupported type")))
					Expect(credentials.OtherDetails).To(BeEmpty())
				})
			})

			Describe("GetOtherDetails", func() {
				It("unmarshalls json content", func() {
					serviceBindingCredentials := models.ServiceBindingCredentials{
						OtherDetails: []byte(`{"some":["json","blob","here"]}`),
					}

					var actualOtherDetails map[string]interface{}
					err := serviceBindingCredentials.GetOtherDetails(&actualOtherDetails)

					Expect(err).ToNot(HaveOccurred())
					Expect(actualOtherDetails).To(HaveLen(1))
					Expect(actualOtherDetails).To(HaveKeyWithValue("some", ConsistOf("json", "blob", "here")))
				})

				It("returns nil if is empty", func() {
					serviceBindingCredentials := models.ServiceBindingCredentials{}

					var actualOtherDetails map[string]interface{}
					err := serviceBindingCredentials.GetOtherDetails(&actualOtherDetails)

					Expect(err).ToNot(HaveOccurred())

					Expect(actualOtherDetails).To(BeNil())
				})

				It("returns an error if unmarshalling fails", func() {
					serviceBindingCredentials := models.ServiceBindingCredentials{
						OtherDetails: []byte(`{"some":"badjson","here"]}`),
					}

					var actualOtherDetails map[string]interface{}
					err := serviceBindingCredentials.GetOtherDetails(&actualOtherDetails)

					Expect(err).To(MatchError(ContainSubstring("invalid character")))

					Expect(actualOtherDetails).To(BeNil())
				})
			})
		})

		Describe("errors", func() {
			Describe("SetOtherDetails", func() {
				BeforeEach(func() {
					fakeEncryptor := &fakes.FakeEncryptor{}
					fakeEncryptor.EncryptReturns(nil, errors.New("fake encrypt error"))

					encryptor = fakeEncryptor
					models.SetEncryptor(encryptor)
				})

				It("returns the error", func() {
					credentials := models.ServiceBindingCredentials{}
					err := credentials.SetOtherDetails("foo")
					Expect(err).To(MatchError("fake encrypt error"))
				})
			})

			Describe("GetOtherDetails", func() {
				BeforeEach(func() {
					fakeEncryptor := &fakes.FakeEncryptor{}
					fakeEncryptor.DecryptReturns(nil, errors.New("fake decrypt error"))

					encryptor = fakeEncryptor
					models.SetEncryptor(encryptor)
				})

				It("return the error", func() {
					serviceBindingCredentials := models.ServiceBindingCredentials{
						OtherDetails: []byte("fake stuff"),
					}

					var receiver interface{}
					err := serviceBindingCredentials.GetOtherDetails(&receiver)

					Expect(err).To(MatchError("fake decrypt error"))
				})
			})
		})
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

	Describe("ProvisionRequestDetails", func() {
		Context("GCM encryptor", func() {
			BeforeEach(func() {
				key := newKey()
				encryptor = gcmencryptor.New(key)
				models.SetEncryptor(encryptor)
			})

			Describe("SetRequestDetails", func() {
				It("encrypts and sets the details", func() {
					details := models.ProvisionRequestDetails{}

					rawMessage := []byte(`{"key":"value"}`)
					details.SetRequestDetails(rawMessage)

					decryptedDetails, err := encryptor.Decrypt(details.RequestDetails)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(decryptedDetails)).To(Equal(`{"key":"value"}`))
				})

				It("converts nil to the empty string", func() {
					details := models.ProvisionRequestDetails{}

					details.SetRequestDetails(nil)

					decryptedDetails, err := encryptor.Decrypt(details.RequestDetails)
					Expect(err).ToNot(HaveOccurred())
					Expect(decryptedDetails).To(BeEmpty())
				})

				It("converts empty array to the empty string", func() {
					details := models.ProvisionRequestDetails{}
					var rawMessage []byte
					details.SetRequestDetails(rawMessage)

					decryptedDetails, err := encryptor.Decrypt(details.RequestDetails)
					Expect(err).ToNot(HaveOccurred())
					Expect(decryptedDetails).To(BeEmpty())
				})
			})

			Describe("GetRequestDetails", func() {
				It("gets as RawMessage", func() {
					encryptedDetails, _ := encryptor.Encrypt([]byte(`{"some":["json","blob","here"]}`))
					requestDetails := models.ProvisionRequestDetails{
						RequestDetails: encryptedDetails,
					}

					details, err := requestDetails.GetRequestDetails()

					rawMessage := json.RawMessage(`{"some":["json","blob","here"]}`)

					Expect(err).ToNot(HaveOccurred())
					Expect(details).To(Equal(rawMessage))
				})
			})

			It("Can decrypt what it had previously encrypted", func() {
				details := models.ProvisionRequestDetails{}

				rawMessage := json.RawMessage(`{"key":"value"}`)
				details.SetRequestDetails(rawMessage)

				actualDetails, err := details.GetRequestDetails()

				Expect(err).ToNot(HaveOccurred())
				Expect(actualDetails).To(Equal(rawMessage))
			})
		})

		Context("Noop encryptor", func() {
			BeforeEach(func() {
				encryptor = noopencryptor.New()
				models.SetEncryptor(encryptor)
			})

			Describe("SetRequestDetails", func() {
				It("sets the details", func() {
					details := models.ProvisionRequestDetails{}

					rawMessage := []byte(`{"key":"value"}`)
					details.SetRequestDetails(rawMessage)

					Expect(details.RequestDetails).To(Equal(rawMessage))
				})

				It("converts nil to the empty string", func() {
					details := models.ProvisionRequestDetails{}

					details.SetRequestDetails(nil)

					Expect(details.RequestDetails).To(BeNil())
				})

				It("converts empty array to the empty string", func() {
					details := models.ProvisionRequestDetails{}
					var rawMessage []byte
					details.SetRequestDetails(rawMessage)

					Expect(details.RequestDetails).To(BeEmpty())
				})
			})

			Describe("GetRequestDetails", func() {
				It("gets as RawMessage", func() {
					requestDetails := models.ProvisionRequestDetails{
						RequestDetails: []byte(`{"some":["json","blob","here"]}`),
					}

					details, err := requestDetails.GetRequestDetails()

					rawMessage := json.RawMessage(`{"some":["json","blob","here"]}`)

					Expect(err).ToNot(HaveOccurred())
					Expect(details).To(Equal(rawMessage))
				})
			})
		})

		Describe("errors", func() {
			Context("SetRequestDetails", func() {
				BeforeEach(func() {
					fakeEncryptor := &fakes.FakeEncryptor{}
					fakeEncryptor.EncryptReturns([]byte{}, errors.New("some error"))

					encryptor = fakeEncryptor
					models.SetEncryptor(encryptor)
				})

				It("returns an error when there are errors while encrypting", func() {
					details := models.ProvisionRequestDetails{}
					var rawMessage []byte

					err := details.SetRequestDetails(rawMessage)

					Expect(err).To(MatchError("some error"))
				})
			})

			Context("GetRequestDetails", func() {
				BeforeEach(func() {
					fakeEncryptor := &fakes.FakeEncryptor{}
					fakeEncryptor.DecryptReturns(nil, errors.New("some error"))

					encryptor = fakeEncryptor
					models.SetEncryptor(encryptor)
				})

				It("returns an error when there are errors while decrypting", func() {
					requestDetails := models.ProvisionRequestDetails{
						RequestDetails: []byte("some string"),
					}

					details, err := requestDetails.GetRequestDetails()

					Expect(err).To(MatchError("some error"))
					Expect(details).To(BeNil())

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
