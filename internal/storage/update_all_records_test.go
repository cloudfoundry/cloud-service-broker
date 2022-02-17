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
			Expect(receiver[0].Workspace).To(Equal([]byte(`{"encrypted":{"decrypted":fake-workspace-1}}`)))
			Expect(receiver[1].Workspace).To(Equal([]byte(`{"encrypted":{"decrypted":fake-workspace-2}}`)))
			Expect(receiver[2].Workspace).To(Equal([]byte(`{"encrypted":{"decrypted":fake-workspace-3}}`)))
		})
	})
})
