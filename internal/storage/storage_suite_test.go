package storage_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/storage"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/storage/storagefakes"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"gorm.io/driver/sqlite"

	"gorm.io/gorm"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	encryptor = &storagefakes.FakeEncryptor{
		DecryptStub: func(bytes []byte) ([]byte, error) {
			return []byte(`{"decrypted":` + string(bytes) + `}`), nil
		},
		EncryptStub: func(bytes []byte) ([]byte, error) {
			return []byte(`{"encrypted":` + string(bytes) + `}`), nil
		},
	}

	store = storage.New(db, encryptor)
})
