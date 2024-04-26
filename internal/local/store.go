package local

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/cloudfoundry/cloud-service-broker/v3/dbservice"
	"github.com/cloudfoundry/cloud-service-broker/v3/internal/encryption/noopencryptor"
	"github.com/cloudfoundry/cloud-service-broker/v3/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v3/internal/testdrive"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var once sync.Once

func store() *storage.Storage {
	dbConn, err := gorm.Open(sqlite.Open(databasePath()), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	ensureDatabaseTablesExist(dbConn)
	return storage.New(dbConn, noopencryptor.New())
}

func ensureDatabaseTablesExist(dbConn *gorm.DB) {
	once.Do(func() {
		if err := dbservice.RunMigrations(dbConn); err != nil {
			log.Fatalf("Error migrating database: %s", err)
		}
	})
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
