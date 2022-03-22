package storage_test

import (
	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpdateAllRecords", func() {

	BeforeEach(func() {
		addFakeServiceCredentialBindings()
		addFakeProvisionRequestDetails()
		addFakeBindRequestDetails()
		addFakeServiceInstanceDetails()
		addFakeTerraformDeployments()
	})

	It("updates all the records with the latest encoding", func() {
		Expect(store.UpdateAllRecords()).NotTo(HaveOccurred())

		By("checking service binding credentials", func() {
			var receiver []models.ServiceBindingCredentials
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver).To(HaveLen(3))
			Expect(receiver[0].OtherDetails).To(MatchJSON(`{"encrypted":{"decrypted":{"foo":"bar"}}}`))
			Expect(receiver[1].OtherDetails).To(MatchJSON(`{"encrypted":{"decrypted":{"foo":"baz","bar":"quz"}}}`))
			Expect(receiver[2].OtherDetails).To(MatchJSON(`{"encrypted":{"decrypted":{"foo":"boz"}}}`))
		})

		By("checking bind request details", func() {
			var receiver []models.BindRequestDetails
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver).To(HaveLen(3))
			Expect(receiver[0].RequestDetails).To(Equal([]byte(`{"encrypted":{"decrypted":{"foo":"bar"}}}`)))
			Expect(receiver[1].RequestDetails).To(Equal([]byte(`{"encrypted":{"decrypted":{"foo":"baz","bar":"quz"}}}`)))
			Expect(receiver[2].RequestDetails).To(Equal([]byte(`{"encrypted":{"decrypted":{"foo":"boz"}}}`)))
		})

		By("checking provision request details", func() {
			var receiver []models.ProvisionRequestDetails
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver).To(HaveLen(3))
			Expect(receiver[0].RequestDetails).To(Equal([]byte(`{"encrypted":{"decrypted":{"foo":"bar"}}}`)))
			Expect(receiver[1].RequestDetails).To(Equal([]byte(`{"encrypted":{"decrypted":{"foo":"baz","bar":"quz"}}}`)))
			Expect(receiver[2].RequestDetails).To(Equal([]byte(`{"encrypted":{"decrypted":{"foo":"boz"}}}`)))
		})

		By("checking service instance details", func() {
			var receiver []models.ServiceInstanceDetails
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver).To(HaveLen(3))
			Expect(receiver[0].OtherDetails).To(Equal([]byte(`{"encrypted":{"decrypted":{"foo":"bar-1"}}}`)))
			Expect(receiver[1].OtherDetails).To(Equal([]byte(`{"encrypted":{"decrypted":{"foo":"bar-2"}}}`)))
			Expect(receiver[2].OtherDetails).To(Equal([]byte(`{"encrypted":{"decrypted":{"foo":"bar-3"}}}`)))
		})

		By("checking terraform deployments", func() {
			var receiver []models.TerraformDeployment
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver).To(HaveLen(3))
			Expect(receiver[0].Workspace).To(Equal([]byte(`{"encrypted":{"decrypted":{"workspace":"fake-1"}}}`)))
			Expect(receiver[1].Workspace).To(Equal([]byte(`{"encrypted":{"decrypted":{"workspace":"fake-2"}}}`)))
			Expect(receiver[2].Workspace).To(Equal([]byte(`{"encrypted":{"decrypted":{"workspace":"fake-3"}}}`)))
		})
	})

	Describe("errors", func() {
		Context("service binding credentials", func() {
			When("OtherDetails cannot be decrypted", func() {
				BeforeEach(func() {
					Expect(db.Create(&models.ServiceBindingCredentials{
						OtherDetails:      []byte(`cannot-be-decrypted`),
						ServiceId:         "fake-other-service-id",
						ServiceInstanceId: "fake-instance-id",
						BindingId:         "fake-bad-binding-id",
					}).Error).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(store.UpdateAllRecords()).To(MatchError(`error re-encoding service binding credentials: decode error for "fake-bad-binding-id": decryption error: fake decryption error`))
				})
			})

			When("OtherDetails cannot be encrypted", func() {
				BeforeEach(func() {
					Expect(db.Create(&models.ServiceBindingCredentials{
						OtherDetails:      []byte(`"cannot-be-encrypted"`),
						ServiceId:         "fake-other-service-id",
						ServiceInstanceId: "fake-instance-id",
						BindingId:         "fake-bad-binding-id",
					}).Error).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(store.UpdateAllRecords()).To(MatchError(`error re-encoding service binding credentials: encode error for "fake-bad-binding-id": encryption error: fake encryption error`))
				})
			})
		})

		Context("service binding request details", func() {
			When("RequestDetails cannot be decrypted", func() {
				BeforeEach(func() {
					Expect(db.Create(&models.BindRequestDetails{
						RequestDetails:    []byte(`cannot-be-decrypted`),
						ServiceBindingId:  "fake-bad-binding-id",
						ServiceInstanceId: "fake-instance-id",
					}).Error).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(store.UpdateAllRecords()).To(MatchError(`error re-encoding service binding request details: decode error for "fake-bad-binding-id": decryption error: fake decryption error`))
				})
			})

			When("RequestDetails cannot be encrypted", func() {
				BeforeEach(func() {
					Expect(db.Create(&models.BindRequestDetails{
						RequestDetails:    []byte(`cannot-be-encrypted`),
						ServiceBindingId:  "fake-bad-binding-id",
						ServiceInstanceId: "fake-instance-id",
					}).Error).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(store.UpdateAllRecords()).To(MatchError(`error re-encoding service binding request details: encode error for "fake-bad-binding-id": encryption error: fake encryption error`))
				})
			})
		})

		Context("provision request details", func() {
			When("RequestDetails cannot be decrypted", func() {
				BeforeEach(func() {
					Expect(db.Create(&models.ProvisionRequestDetails{
						RequestDetails:    []byte(`cannot-be-decrypted`),
						ServiceInstanceId: "fake-bad-instance-id",
					}).Error).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(store.UpdateAllRecords()).To(MatchError(`error re-encoding provision request details: decode error for "fake-bad-instance-id": decryption error: fake decryption error`))
				})
			})

			When("RequestDetails cannot be encrypted", func() {
				BeforeEach(func() {
					Expect(db.Create(&models.ProvisionRequestDetails{
						RequestDetails:    []byte(`cannot-be-encrypted`),
						ServiceInstanceId: "fake-bad-instance-id",
					}).Error).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(store.UpdateAllRecords()).To(MatchError(`error re-encoding provision request details: encode error for "fake-bad-instance-id": encryption error: fake encryption error`))
				})
			})
		})

		Context("service instance details", func() {
			When("OtherDetails cannot be decrypted", func() {
				BeforeEach(func() {
					Expect(db.Create(&models.ServiceInstanceDetails{
						ID:           "fake-bad-id",
						OtherDetails: []byte(`cannot-be-decrypted`),
					}).Error).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(store.UpdateAllRecords()).To(MatchError(`error re-encoding service instance details: decode error for "fake-bad-id": decryption error: fake decryption error`))
				})
			})

			When("OtherDetails cannot be encrypted", func() {
				BeforeEach(func() {
					Expect(db.Create(&models.ServiceInstanceDetails{
						ID:           "fake-bad-id",
						OtherDetails: []byte(`"cannot-be-encrypted"`),
					}).Error).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(store.UpdateAllRecords()).To(MatchError(`error re-encoding service instance details: encode error for "fake-bad-id": encryption error: fake encryption error`))
				})
			})
		})

		Context("terraform deployments", func() {
			When("Workspace cannot be decrypted", func() {
				BeforeEach(func() {
					Expect(db.Create(&models.TerraformDeployment{
						ID:        "fake-bad-id",
						Workspace: []byte("cannot-be-decrypted"),
					}).Error).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(store.UpdateAllRecords()).To(MatchError(`error re-encoding terraform deployment: decode error for "fake-bad-id": decryption error: fake decryption error`))
				})
			})

			When("Workspace cannot be encrypted", func() {
				BeforeEach(func() {
					Expect(db.Create(&models.TerraformDeployment{
						ID:        "fake-bad-id",
						Workspace: []byte("cannot-be-encrypted"),
					}).Error).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(store.UpdateAllRecords()).To(MatchError(`error re-encoding terraform deployment: encode error for "fake-bad-id": encryption error: fake encryption error`))
				})
			})
		})
	})
})
