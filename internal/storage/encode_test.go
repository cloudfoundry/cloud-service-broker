package storage_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage/storagefakes"
)

var _ = Describe("Encode", func() {
	// These tests verify the fix for handling NULL/empty byte arrays during
	// decryption, which prevents "malformed ciphertext" errors during service
	// instance upgrades (see PR #1341 migration issue).

	Describe("decodeBytes handling of empty input", func() {
		// This tests the scenario where bind_resource is NULL in the database
		// after migration from V1 schema. Previously, attempting to decrypt
		// NULL/empty bytes caused "malformed ciphertext" errors.

		Context("when bind_resource is NULL/empty in database", func() {
			var (
				testEncryptor *storagefakes.FakeEncryptor
				testStore     *storage.Storage
			)

			BeforeEach(func() {
				testEncryptor = &storagefakes.FakeEncryptor{}
				testStore = storage.New(db, testEncryptor)

				// Simulate a record that was migrated from V1 schema where
				// bind_resource column was added but is NULL for existing records
				Expect(db.Create(&models.BindRequestDetails{
					ServiceBindingID:  "binding-with-null-resource",
					ServiceInstanceID: "instance-with-null-resource",
					Parameters:        []byte(`{"foo":"bar"}`),
					BindResource:      nil, // NULL - simulates migrated record
				}).Error).NotTo(HaveOccurred())
			})

			It("does not call decryptor for NULL bind_resource", func() {
				// Configure encryptor to track calls and return valid data for Parameters
				testEncryptor.DecryptStub = func(data []byte) ([]byte, error) {
					// Should only be called for non-empty Parameters, not for NULL BindResource
					if len(data) == 0 {
						Fail("Decrypt should not be called with empty data")
					}
					return data, nil
				}

				_, err := testStore.GetBindRequestDetails("binding-with-null-resource", "instance-with-null-resource")
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not return malformed ciphertext error", func() {
				// Configure encryptor to return "malformed ciphertext" for empty input
				// This simulates what the real GCM encryptor does
				testEncryptor.DecryptStub = func(data []byte) ([]byte, error) {
					if len(data) == 0 {
						return nil, errors.New("malformed ciphertext")
					}
					return data, nil
				}

				// This should NOT fail - the fix prevents empty data from reaching the decryptor
				_, err := testStore.GetBindRequestDetails("binding-with-null-resource", "instance-with-null-resource")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when bind_resource is empty byte slice in database", func() {
			var (
				testEncryptor *storagefakes.FakeEncryptor
				testStore     *storage.Storage
			)

			BeforeEach(func() {
				testEncryptor = &storagefakes.FakeEncryptor{}
				testStore = storage.New(db, testEncryptor)

				Expect(db.Create(&models.BindRequestDetails{
					ServiceBindingID:  "binding-with-empty-resource",
					ServiceInstanceID: "instance-with-empty-resource",
					Parameters:        []byte(`{"foo":"bar"}`),
					BindResource:      []byte{}, // Empty slice
				}).Error).NotTo(HaveOccurred())
			})

			It("does not call decryptor for empty bind_resource", func() {
				testEncryptor.DecryptStub = func(data []byte) ([]byte, error) {
					if len(data) == 0 {
						Fail("Decrypt should not be called with empty data")
					}
					return data, nil
				}

				_, err := testStore.GetBindRequestDetails("binding-with-empty-resource", "instance-with-empty-resource")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when bind_resource has valid encrypted data", func() {
			var (
				testEncryptor *storagefakes.FakeEncryptor
				testStore     *storage.Storage
			)

			BeforeEach(func() {
				testEncryptor = &storagefakes.FakeEncryptor{}
				testStore = storage.New(db, testEncryptor)

				Expect(db.Create(&models.BindRequestDetails{
					ServiceBindingID:  "binding-with-valid-resource",
					ServiceInstanceID: "instance-with-valid-resource",
					Parameters:        []byte(`{"foo":"bar"}`),
					BindResource:      []byte(`{"app_guid":"test-app"}`),
				}).Error).NotTo(HaveOccurred())
			})

			It("calls decryptor for non-empty bind_resource", func() {
				testEncryptor.DecryptStub = func(data []byte) ([]byte, error) {
					return data, nil
				}

				_, err := testStore.GetBindRequestDetails("binding-with-valid-resource", "instance-with-valid-resource")
				Expect(err).NotTo(HaveOccurred())

				// Should be called twice: once for Parameters, once for BindResource
				Expect(testEncryptor.DecryptCallCount()).To(Equal(2))
			})

			When("decryption fails", func() {
				It("returns a wrapped decryption error", func() {
					testEncryptor.DecryptReturns(nil, errors.New("malformed ciphertext"))

					_, err := testStore.GetBindRequestDetails("binding-with-valid-resource", "instance-with-valid-resource")
					Expect(err).To(MatchError(ContainSubstring("decryption error: malformed ciphertext")))
				})
			})
		})
	})
})
