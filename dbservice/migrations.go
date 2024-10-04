// Copyright 2018 the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dbservice

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
)

const numMigrations = 17

// RunMigrations runs schema migrations on the provided service broker database to get it up to date
func RunMigrations(db *gorm.DB) error {
	migrations := make([]func() error, numMigrations)

	// initial migration - creates tables
	migrations[0] = func() error { // v1.0
		return autoMigrateTables(db,
			&models.ServiceInstanceDetailsV1{},
			&models.ServiceBindingCredentialsV1{},
			&models.ProvisionRequestDetailsV1{},
			&models.PlanDetailsV1{},
			&models.MigrationV1{})
	}

	// adds CloudOperation table
	migrations[1] = func() error { // v2.x
		// NOTE: this migration used to have lots of custom logic, however it has
		// been removed because brokers starting at v4 no longer support the
		// functionality the migration required.
		//
		// It is acceptable to pass through this migration step on the way to
		// initialize a _new_ database, but it is not acceptable to use this step
		// in a path through the upgrade.
		return autoMigrateTables(db, &models.CloudOperationV1{})
	}

	// drops plan details table
	migrations[2] = func() error { // 4.0.0
		// NOOP migration, this was used to drop the plan_details table, but
		// there's more of a disincentive than incentive to do that because it could
		// leave operators wiping out plain details accidentally and not being able
		// to recover if they don't follow the upgrade path.
		return nil
	}

	migrations[3] = func() error { // v4.1.0
		return autoMigrateTables(db, &models.ServiceInstanceDetailsV2{})
	}

	migrations[4] = func() error { // v4.2.0
		return autoMigrateTables(db, &models.TerraformDeploymentV1{})
	}

	migrations[5] = func() error { // v4.2.3
		return autoMigrateTables(db, &models.ProvisionRequestDetailsV2{})
	}

	migrations[6] = func() error { // v4.2.4
		if db.Config.Dialector.Name() == "sqlite3" {
			// sqlite does not support changing column data types
			return nil
		} else {
			return db.Migrator().AlterColumn(&models.ProvisionRequestDetailsV2{}, "request_details")
		}
	}

	migrations[7] = func() error { // v0.2.2
		return autoMigrateTables(db, &models.TerraformDeploymentV2{})
	}

	migrations[8] = func() error { // v0.2.2
		if db.Config.Dialector.Name() == "sqlite3" {
			// sqlite does not support changing column data types.
			// Shouldn't matter because sqlite is only for non-prod deployments,
			// and can be re-provisioned more easily.
			return nil
		} else {
			return db.Migrator().AlterColumn(&models.TerraformDeploymentV2{}, "workspace")
		}
	}

	migrations[9] = func() error {
		return autoMigrateTables(db, &models.PasswordMetadataV1{})
	}

	migrations[10] = func() error {
		return db.Migrator().AlterColumn(&models.ProvisionRequestDetailsV3{}, "request_details")
	}

	migrations[11] = func() error {
		return db.Migrator().AlterColumn(&models.ServiceInstanceDetailsV3{}, "other_details")
	}

	migrations[12] = func() error {
		return db.Migrator().AlterColumn(&models.ServiceBindingCredentialsV2{}, "other_details")
	}

	migrations[13] = func() error {
		// This used to be a migration step that altered TerraformDeployment workspace field type from mediumtext to blob.
		// That resulted in decreased field capacity (16384K to 64K).
		// In order to keep the right number of migrations and fix the issue we should keep this migration id and add
		// one more to update to larger field size.
		return nil
	}

	migrations[14] = func() error {
		return db.Migrator().AlterColumn(&models.TerraformDeploymentV3{}, "workspace")
	}

	migrations[15] = func() error {
		return autoMigrateTables(db, &models.BindRequestDetailsV1{})
	}

	migrations[16] = func() error {
		if err := db.Migrator().DropColumn(&models.ServiceInstanceDetailsV4{}, "operation_type"); err != nil {
			return err
		}
		if err := db.Migrator().DropColumn(&models.ServiceInstanceDetailsV4{}, "operation_id"); err != nil {
			return err
		}
		if err := db.Migrator().DropColumn(&models.ServiceInstanceDetailsV4{}, "location"); err != nil {
			return err
		}
		return db.Migrator().DropColumn(&models.ServiceInstanceDetailsV4{}, "url")
	}

	var lastMigrationNumber = -1

	// if we've run any migrations before, we should have a migrations table, so find the last one we ran
	if db.Migrator().HasTable("migrations") {
		var storedMigrations []models.Migration
		if err := db.Order("migration_id desc").Find(&storedMigrations).Error; err != nil {
			return fmt.Errorf("error getting last migration id even though migration table exists: %s", err)
		}
		lastMigrationNumber = storedMigrations[0].MigrationID
	}

	if err := ValidateLastMigration(lastMigrationNumber); err != nil {
		return err
	}

	// starting from the last migration we ran + 1, run migrations until we are current
	for i := lastMigrationNumber + 1; i < len(migrations); i++ {
		tx := db.Begin()
		err := migrations[i]()
		if err != nil {
			tx.Rollback()

			return err
		} else {
			newMigration := models.Migration{
				MigrationID: i,
			}
			if err := db.Save(&newMigration).Error; err != nil {
				tx.Rollback()
				return err
			} else {
				tx.Commit()
			}
		}
	}

	return nil
}

// ValidateLastMigration returns an error if the database version is newer than
// this tool supports or is too old to be updated.
func ValidateLastMigration(lastMigration int) error {
	switch {
	case lastMigration >= numMigrations:
		return errors.New("the database you're connected to is newer than this tool supports")

	case lastMigration == 0:
		return errors.New("migration from broker versions <= 2.0 is no longer supported, upgrade using a v3.x broker then try again")

	default:
		return nil
	}
}

func autoMigrateTables(db *gorm.DB, tables ...any) error {
	if db.Config.Dialector.Name() == "mysql" {
		return db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8").AutoMigrate(tables...)
	} else {
		return db.AutoMigrate(tables...)
	}
}
