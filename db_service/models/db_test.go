package models_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Db", func() {
	BeforeEach(func() {
		models.NewKey()
	})

	Describe("ServiceBindingCredentials", func() {
		Describe("SetOtherDetails", func() {
			It("marshalls json content", func() {
				var arrayOfInterface []interface{}
				arrayOfInterface = append(arrayOfInterface, "json", "blob", "here")
				otherDetails := map[string]interface{}{
					"some": arrayOfInterface,
				}
				credentials := models.ServiceBindingCredentials{}

				err := credentials.SetOtherDetails(otherDetails)

				Expect(err).To(BeNil(), "Should not have returned an error")
				Expect(credentials.OtherDetails).To(Equal("{\"some\":[\"json\",\"blob\",\"here\"]}"))
			})

			It("marshalls nil into json null", func() {
				credentials := models.ServiceBindingCredentials{}

				err := credentials.SetOtherDetails(nil)

				Expect(err).To(BeNil(), "Should not have returned an error")
				Expect(credentials.OtherDetails).To(Equal("null"))
			})

			It("returns an error if it cannot marshall", func() {

				credentials := models.ServiceBindingCredentials{}

				err := credentials.SetOtherDetails(struct {
					F func()
				}{F: func() {}})

				Expect(err).ToNot(BeNil(), "Should have returned an error")
				Expect(credentials.OtherDetails).To(BeEmpty())
			})
		})

		Describe("GetOtherDetails", func() {
			It("unmarshalls json content", func() {
				serviceBindingCredentials := models.ServiceBindingCredentials{
					OtherDetails: "{\"some\":[\"json\",\"blob\",\"here\"]}",
				}

				var actualOtherDetails map[string]interface{}
				err := serviceBindingCredentials.GetOtherDetails(&actualOtherDetails)

				Expect(err).To(BeNil(), "Should not have returned an error")

				var arrayOfInterface []interface{}
				arrayOfInterface = append(arrayOfInterface, "json", "blob", "here")
				expectedOtherDetails := map[string]interface{}{
					"some": arrayOfInterface,
				}
				Expect(actualOtherDetails).To(Equal(expectedOtherDetails))
			})

			It("returns nil if is empty", func() {
				serviceBindingCredentials := models.ServiceBindingCredentials{}

				var actualOtherDetails map[string]interface{}
				err := serviceBindingCredentials.GetOtherDetails(&actualOtherDetails)

				Expect(err).To(BeNil(), "Should not have returned an error")

				Expect(actualOtherDetails).To(BeNil())
			})

			It("returns an error if unmarshalling fails", func() {
				serviceBindingCredentials := models.ServiceBindingCredentials{
					OtherDetails: "{\"some\":\"badjson\", \"here\"]}",
				}

				var actualOtherDetails map[string]interface{}
				err := serviceBindingCredentials.GetOtherDetails(&actualOtherDetails)

				Expect(err).To(MatchError(ContainSubstring("invalid character")))

				Expect(actualOtherDetails).To(BeNil())
			})
		})

		It("Can decrypt what it had previously encrypted", func() {
			var arrayOfInterface []interface{}
			arrayOfInterface = append(arrayOfInterface, "json", "blob", "here")
			otherDetails := map[string]interface{}{
				"some": arrayOfInterface,
			}
			credentials := models.ServiceBindingCredentials{}

			credentials.SetOtherDetails(otherDetails)

			var actualOtherDetails map[string]interface{}
			err := credentials.GetOtherDetails(&actualOtherDetails)

			Expect(err).To(BeNil(), "Should not have returned an error")

			Expect(actualOtherDetails).To(Equal(otherDetails))
		})
	})

	Describe("ServiceInstanceDetails", func() {
		Describe("SetOtherDetails", func() {
			It("marshalls and encrypts json content", func() {
				var arrayOfInterface []interface{}
				arrayOfInterface = append(arrayOfInterface, "json", "blob", "here")
				otherDetails := map[string]interface{}{
					"some": arrayOfInterface,
				}
				details := models.ServiceInstanceDetails{}

				err := details.SetOtherDetails(otherDetails)

				Expect(err).To(BeNil(), "Should not have returned an error")
				decryptedDetails, _ := models.Decrypt([]byte(details.OtherDetails), &models.Key)
				Expect(string(decryptedDetails)).To(Equal("{\"some\":[\"json\",\"blob\",\"here\"]}"))
			})

			It("marshalls nil into json null", func() {
				details := models.ServiceInstanceDetails{}

				err := details.SetOtherDetails(nil)

				Expect(err).To(BeNil(), "Should not have returned an error")
				decryptedDetails, _ := models.Decrypt([]byte(details.OtherDetails), &models.Key)
				Expect(string(decryptedDetails)).To(Equal("null"))
			})

			It("returns an error if it cannot marshall", func() {

				details := models.ServiceInstanceDetails{}

				err := details.SetOtherDetails(struct {
					F func()
				}{F: func() {}})

				Expect(err).ToNot(BeNil(), "Should have returned an error")
				Expect(details.OtherDetails).To(BeEmpty())
			})
		})

		Describe("GetOtherDetails", func() {
			It("decrypts and unmarshalls json content", func() {
				encryptedDetails, _ := models.Encrypt([]byte("{\"some\":[\"json\",\"blob\",\"here\"]}"), &models.Key)
				serviceInstanceDetails := models.ServiceInstanceDetails{
					OtherDetails: string(encryptedDetails),
				}

				var actualOtherDetails map[string]interface{}
				err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

				Expect(err).To(BeNil(), "Should not have returned an error")

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

				Expect(err).To(BeNil(), "Should not have returned an error")

				Expect(actualOtherDetails).To(BeNil())
			})

			It("returns an error if unmarshalling fails", func() {
				encryptedDetails, _ := models.Encrypt([]byte("{\"some\":\"badjson\", \"here\"]}"), &models.Key)
				serviceInstanceDetails := models.ServiceInstanceDetails{
					OtherDetails: string(encryptedDetails),
				}

				var actualOtherDetails map[string]interface{}
				err := serviceInstanceDetails.GetOtherDetails(&actualOtherDetails)

				Expect(err).To(MatchError(ContainSubstring("invalid character")))

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

			Expect(err).To(BeNil(), "Should not have returned an error")

			var arrayOfInterface []interface{}
			arrayOfInterface = append(arrayOfInterface, "json", "blob", "here")
			expectedOtherDetails := map[string]interface{}{
				"some": arrayOfInterface,
			}

			Expect(actualOtherDetails).To(Equal(expectedOtherDetails))
		})
	})

	Describe("ProvisionRequestDetails", func() {
		Describe("SetRequestDetails", func() {
			It("sets the details as string value", func() {
				details := models.ProvisionRequestDetails{}

				rawMessage := []byte(`{"key":"value"}`)
				details.SetRequestDetails(rawMessage)

				Expect(details.RequestDetails).To(Equal("{\"key\":\"value\"}"))
			})

			It("converts nil to the empty string", func() {
				details := models.ProvisionRequestDetails{}

				details.SetRequestDetails(nil)

				Expect(details.RequestDetails).To(BeEmpty())
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
					RequestDetails: "{\"some\":[\"json\",\"blob\",\"here\"]}",
				}

				details := requestDetails.GetRequestDetails()

				rawMessage := json.RawMessage([]byte(`{"some":["json","blob","here"]}`))

				Expect(details).To(Equal(rawMessage))
			})
		})
	})
})
