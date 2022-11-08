package local

import (
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cloud-service-broker/internal/encryption/noopencryptor"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func store() *storage.Storage {
	dbConn, err := gorm.Open(sqlite.Open(databasePath()), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	return storage.New(dbConn, noopencryptor.New())
}

func databasePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(cwd, ".csb.db")
}

func lookupServiceInstanceByGUID(guid string) testdrive.ServiceInstance {
	d, err := store().GetServiceInstanceDetails(guid)
	if err != nil {
		log.Fatal(err)
	}

	return testdrive.ServiceInstance{
		GUID:                guid,
		ServicePlanGUID:     d.PlanGUID,
		ServiceOfferingGUID: d.ServiceGUID,
	}
}
