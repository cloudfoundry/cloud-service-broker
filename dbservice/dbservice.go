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

// Package dbservice implements database setup, connection, and migration
package dbservice

import (
	"fmt"
	"sync"

	"code.cloudfoundry.org/lager"
	"gorm.io/gorm"

	_ "gorm.io/driver/sqlite"
)

var once sync.Once

// New instantiates the db connection and runs migrations
func New(logger lager.Logger) *gorm.DB {
	var db *gorm.DB
	once.Do(func() {
		db = SetupDB(logger)
		if err := RunMigrations(db); err != nil {
			panic(fmt.Sprintf("Error migrating database: %s", err.Error()))
		}
	})
	return db
}
