package storage_test

import (
	"errors"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage/storagefakes"
)

var (
	db        *gorm.DB
	encryptor *storagefakes.FakeEncryptor
	store     *storage.Storage
)

func TestStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Suite")
}

var _ = BeforeEach(func() {
	var err error
	db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Migrator().CreateTable(&models.ServiceBindingCredentials{})).NotTo(HaveOccurred())
	Expect(db.Migrator().CreateTable(&models.ProvisionRequestDetails{})).NotTo(HaveOccurred())
	Expect(db.Migrator().CreateTable(&models.BindRequestDetails{})).NotTo(HaveOccurred())
	Expect(db.Migrator().CreateTable(&models.ServiceInstanceDetails{})).NotTo(HaveOccurred())
	Expect(db.Migrator().CreateTable(&models.TerraformDeployment{})).NotTo(HaveOccurred())

	encryptor = &storagefakes.FakeEncryptor{
		DecryptStub: func(bytes []byte) ([]byte, error) {
			if string(bytes) == `cannot-be-decrypted` {
				return nil, errors.New("fake decryption error")
			}
			return []byte(`{"decrypted":` + string(bytes) + `}`), nil
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
