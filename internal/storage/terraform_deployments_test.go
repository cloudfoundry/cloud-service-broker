package storage_test

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage/storagefakes"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/workspace"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TerraformDeployments", func() {

	BeforeEach(func() {
		By("overriding the default FakeEncryptor to not change the json on decryption")
		encryptor = &storagefakes.FakeEncryptor{
			DecryptStub: func(bytes []byte) ([]byte, error) {
				if string(bytes) == `cannot-be-decrypted` {
					return nil, errors.New("fake decryption error")
				}
				return bytes, nil
			},
			EncryptStub: func(bytes []byte) ([]byte, error) {
				if strings.Contains(string(bytes), `cannot-be-encrypted`) {
					return nil, errors.New("fake encryption error")
				}
				return []byte(`{"encrypted":` + string(bytes) + `}`), nil
			},
		}

		store = storage.New(db, encryptor)
	})

	Describe("StoreTerraformDeployments", func() {
		It("creates the right object in the database", func() {
			err := store.StoreTerraformDeployment(storage.TerraformDeployment{
				ID: "fake-id",
				Workspace: &workspace.TerraformWorkspace{
					Modules: []workspace.ModuleDefinition{
						{
							Name: "first",
						},
					},
				},
				LastOperationType:    "create",
				LastOperationState:   "succeeded",
				LastOperationMessage: "yes!!",
			})
			Expect(err).NotTo(HaveOccurred())

			var receiver models.TerraformDeployment
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver.ID).To(Equal("fake-id"))
			Expect(receiver.Workspace).To(Equal([]byte(`{"encrypted":{"modules":[{"Name":"first","Definition":"","Definitions":null}],"instances":null,"tfstate":null,"transform":{"parameter_mappings":null,"parameters_to_remove":null,"parameters_to_add":null}}}`)))
			Expect(receiver.LastOperationType).To(Equal("create"))
			Expect(receiver.LastOperationState).To(Equal("succeeded"))
			Expect(receiver.LastOperationMessage).To(Equal("yes!!"))
		})

		When("encoding fails", func() {
			It("returns an error", func() {
				encryptor.EncryptReturns(nil, errors.New("bang"))

				err := store.StoreTerraformDeployment(storage.TerraformDeployment{})
				Expect(err).To(MatchError("error encoding workspace: encryption error: bang"))
			})
		})

		When("details for the instance already exist in the database", func() {
			BeforeEach(func() {
				addFakeTerraformDeployments()
			})

			It("updates the existing record", func() {
				err := store.StoreTerraformDeployment(storage.TerraformDeployment{
					ID: "fake-id-2",
					Workspace: &workspace.TerraformWorkspace{
						Modules: []workspace.ModuleDefinition{
							{
								Name: "first",
							},
						},
					},
					LastOperationType:    "create",
					LastOperationState:   "succeeded",
					LastOperationMessage: "yes!!",
				})
				Expect(err).NotTo(HaveOccurred())

				var receiver models.TerraformDeployment
				Expect(db.Where(`id = "fake-id-2"`).Find(&receiver).Error).NotTo(HaveOccurred())
				Expect(receiver.ID).To(Equal("fake-id-2"))
				Expect(receiver.Workspace).To(Equal([]byte(`{"encrypted":{"modules":[{"Name":"first","Definition":"","Definitions":null}],"instances":null,"tfstate":null,"transform":{"parameter_mappings":null,"parameters_to_remove":null,"parameters_to_add":null}}}`)))
				Expect(receiver.LastOperationType).To(Equal("create"))
				Expect(receiver.LastOperationState).To(Equal("succeeded"))
				Expect(receiver.LastOperationMessage).To(Equal("yes!!"))
			})
		})
	})

	Describe("GetTerraformDeployment", func() {
		BeforeEach(func() {
			addFakeTerraformDeployments()
		})

		It("reads the right object from the database", func() {
			r, err := store.GetTerraformDeployment("fake-id-2")
			Expect(err).NotTo(HaveOccurred())

			Expect(r.ID).To(Equal("fake-id-2"))
			Expect(r.Workspace).To(Equal(&workspace.TerraformWorkspace{Modules: []workspace.ModuleDefinition{{Name: "fake-2"}}}))
			Expect(r.LastOperationType).To(Equal("update"))
			Expect(r.LastOperationState).To(Equal("failed"))
			Expect(r.LastOperationMessage).To(Equal("too bad"))
		})

		When("decoding fails", func() {
			It("returns an error", func() {
				encryptor.DecryptReturns(nil, errors.New("bang"))

				_, err := store.GetTerraformDeployment("fake-id-1")
				Expect(err).To(MatchError(`error decoding workspace "fake-id-1": decryption error: bang`))
			})
		})

		When("nothing is found", func() {
			It("returns an error", func() {
				_, err := store.GetTerraformDeployment("not-there")
				Expect(err).To(MatchError("could not find terraform deployment: not-there"))
			})
		})
	})

	Describe("GetAllTerraformDeployments", func() {
		BeforeEach(func() {
			addFakeTerraformDeployments()
		})

		It("reads all objects from the database", func() {
			r, err := store.GetAllTerraformDeployments()
			Expect(err).NotTo(HaveOccurred())

			Expect(r).To(HaveLen(3))
			Expect(r[0].ID).To(Equal("fake-id-1"))
			Expect(r[0].LastOperationType).To(Equal("create"))
			Expect(r[0].LastOperationState).To(Equal("succeeded"))
			Expect(r[0].LastOperationMessage).To(Equal("amazing"))
			Expect(r[0].StateVersion).To(Equal(version.Must(version.NewVersion("1.2.3"))))
			Expect(r[1].ID).To(Equal("fake-id-2"))
			Expect(r[1].LastOperationType).To(Equal("update"))
			Expect(r[1].LastOperationState).To(Equal("failed"))
			Expect(r[1].LastOperationMessage).To(Equal("too bad"))
			Expect(r[1].StateVersion).To(BeNil())
			Expect(r[2].ID).To(Equal("fake-id-3"))
			Expect(r[2].LastOperationType).To(Equal("update"))
			Expect(r[2].LastOperationState).To(Equal("succeeded"))
			Expect(r[2].LastOperationMessage).To(Equal("great"))
			Expect(r[2].StateVersion).To(Equal(version.Must(version.NewVersion("1.2.4"))))
		})

		When("decoding fails", func() {
			It("returns an error", func() {
				encryptor.DecryptReturns(nil, errors.New("bang"))

				_, err := store.GetAllTerraformDeployments()
				Expect(err).To(MatchError(`error reading terraform deployment batch: error decoding workspace "fake-id-1": decryption error: bang`))
			})
		})
	})

	Describe("ExistsTerraformDeployments", func() {
		BeforeEach(func() {
			addFakeTerraformDeployments()
		})

		It("reads the result from the database", func() {
			Expect(store.ExistsTerraformDeployment("not-there")).To(BeFalse())
			Expect(store.ExistsTerraformDeployment("also-not-there")).To(BeFalse())
			Expect(store.ExistsTerraformDeployment("fake-id-1")).To(BeTrue())
			Expect(store.ExistsTerraformDeployment("fake-id-2")).To(BeTrue())
			Expect(store.ExistsTerraformDeployment("fake-id-3")).To(BeTrue())
		})
	})

	Describe("DeleteTerraformDeployments", func() {
		BeforeEach(func() {
			addFakeTerraformDeployments()
		})

		It("deletes from the database", func() {
			Expect(store.ExistsTerraformDeployment("fake-id-3")).To(BeTrue())

			Expect(store.DeleteTerraformDeployment("fake-id-3")).NotTo(HaveOccurred())

			Expect(store.ExistsTerraformDeployment("fake-id-3")).To(BeFalse())
		})

		It("is idempotent", func() {
			Expect(store.DeleteTerraformDeployment("not-there")).NotTo(HaveOccurred())
		})
	})
})

func addFakeTerraformDeployments() {

	Expect(db.Create(&models.TerraformDeployment{
		ID:                   "fake-id-1",
		Workspace:            fakeWorkspace("fake-1", "1.2.3"),
		LastOperationType:    "create",
		LastOperationState:   "succeeded",
		LastOperationMessage: "amazing",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.TerraformDeployment{
		ID:                   "fake-id-2",
		Workspace:            fakeWorkspace("fake-2", ""),
		LastOperationType:    "update",
		LastOperationState:   "failed",
		LastOperationMessage: "too bad",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.TerraformDeployment{
		ID:                   "fake-id-3",
		Workspace:            fakeWorkspace("fake-3", "1.2.4"),
		LastOperationType:    "update",
		LastOperationState:   "succeeded",
		LastOperationMessage: "great",
	}).Error).NotTo(HaveOccurred())
}

func fakeWorkspace(name, ver string) []byte {
	state := "null"
	if ver != "" {
		state = fmt.Sprintf("%q", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"terraform_version":"%s"}`, ver))))
	}
	return []byte(fmt.Sprintf(`{"modules":[{"Name":"%s","Definition":"","Definitions":null}],"instances":null,"tfstate":%s,"transform":{"parameter_mappings":null,"parameters_to_remove":null,"parameters_to_add":null}}`, name, state))
}

func fakeEncryptedWorkspace(name, ver string) []byte {
	return []byte(fmt.Sprintf(`{"encrypted":{"decrypted":%s}}`, fakeWorkspace(name, ver)))
}
