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

package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/cloudfoundry-incubator/cloud-service-broker/migrator"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/cloud-service-broker/brokerapi/brokers"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/encryption"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/brokerpak"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/server"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/toggles"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils"
	"github.com/gorilla/mux"
	"github.com/pivotal-cf/brokerapi/v8"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const (
	apiUserProp         = "api.user"
	apiPasswordProp     = "api.password"
	apiPortProp         = "api.port"
	apiHostProp         = "api.host"
	encryptionPasswords = "db.encryption.passwords"
	encryptionEnabled   = "db.encryption.enabled"
)

var cfCompatibilityToggle = toggles.Features.Toggle("enable-cf-sharing", false, `Set all services to have the Sharable flag so they can be shared
	across spaces in PCF.`)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start the service broker",
		Long: `Starts the service broker listening on a port defined by the
	PORT environment variable.`,
		Run: func(cmd *cobra.Command, args []string) {
			serve()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "serve-docs",
		Short: "Just serve the docs normally available on the broker",
		Run: func(cmd *cobra.Command, args []string) {
			serveDocs()
		},
	})

	viper.BindEnv(apiUserProp, "SECURITY_USER_NAME")
	viper.BindEnv(apiPasswordProp, "SECURITY_USER_PASSWORD")
	viper.BindEnv(apiPortProp, "PORT")
	viper.BindEnv(apiHostProp, "CSB_LISTENER_HOST")
	viper.BindEnv(encryptionPasswords, "ENCRYPTION_PASSWORDS")
	viper.BindEnv(encryptionEnabled, "ENCRYPTION_ENABLED")
}

func serve() {
	logger := utils.NewLogger("cloud-service-broker")
	db := db_service.New(logger)
	encryptor := setupDBEncryption(db, logger)

	// init broker
	cfg, err := brokers.NewBrokerConfigFromEnv(logger)
	if err != nil {
		logger.Fatal("Error initializing service broker config", err)
	}
	var serviceBroker domain.ServiceBroker
	serviceBroker, err = brokers.New(cfg, logger, storage.New(db, encryptor))
	if err != nil {
		logger.Fatal("Error initializing service broker", err)
	}

	pakConfig, err := brokerpak.NewServerConfigFromEnv()
	if err != nil {
		logger.Fatal("Error initializing broker pack config", err)
	}

	registrar := brokerpak.NewRegistrar(pakConfig)
	migrationRunner, err := migrator.New(cfg, logger, storage.New(db, encryptor), registrar)
	if err != nil {
		logger.Fatal("Error initializing service broker", err)
	}

	credentials := brokerapi.BrokerCredentials{
		Username: viper.GetString(apiUserProp),
		Password: viper.GetString(apiPasswordProp),
	}

	if cfCompatibilityToggle.IsActive() {
		logger.Info("Enabling Cloud Foundry service sharing")
		serviceBroker = server.NewCfSharingWrapper(serviceBroker)
	}

	services, err := serviceBroker.Services(context.Background())
	if err != nil {
		logger.Error("creating service catalog", err)
	}
	logger.Info("service catalog", lager.Data{"catalog": services})

	brokerAPI := brokerapi.New(serviceBroker, logger, credentials)

	sqldb, err := db.DB()
	if err != nil {
		logger.Error("failed to get database connection", err)
	}
	startServer(cfg.Registry, sqldb, brokerAPI, migrationRunner)
}

func serveDocs() {
	logger := utils.NewLogger("cloud-service-broker")
	// init broker
	registry := broker.BrokerRegistry{}
	if err := brokerpak.RegisterAll(registry); err != nil {
		logger.Error("loading brokerpaks", err)
	}

	startServer(registry, nil, nil, nil)
}

func setupDBEncryption(db *gorm.DB, logger lager.Logger) storage.Encryptor {
	config, err := encryption.ParseConfiguration(db, viper.GetBool(encryptionEnabled), viper.GetString(encryptionPasswords))
	if err != nil {
		logger.Fatal("Error parsing encryption configuration", err)
	}

	if config.Changed {
		logger.Info("rotating-database-encryption", lager.Data{"previous-primary": labelName(config.StoredPrimaryLabel), "new-primary": labelName(config.ConfiguredPrimaryLabel)})
		if err := storage.New(db, config.RotationEncryptor).UpdateAllRecords(); err != nil {
			logger.Fatal("Error rotating database encryption", err)
		}
		if err := encryption.UpdatePasswordMetadata(db, config.ConfiguredPrimaryLabel); err != nil {
			logger.Fatal("Error updating password metadata", err)
		}
	}

	if err := encryption.DeletePasswordMetadata(db, config.ToDeleteLabels); err != nil {
		logger.Fatal("Error deleting stale password metadata", err)
	}

	logger.Info("database-encryption", lager.Data{"primary": labelName(config.ConfiguredPrimaryLabel)})
	return config.Encryptor
}

func startServer(registry broker.BrokerRegistry, db *sql.DB, brokerapi http.Handler, runner *migrator.MigrationRunner) {
	logger := utils.NewLogger("cloud-service-broker")

	router := mux.NewRouter()

	// match paths going to the brokerapi first
	if brokerapi != nil {
		router.PathPrefix("/v2").Handler(brokerapi)
	}

	server.AddDocsHandler(router, registry)
	router.HandleFunc("/examples", server.NewExampleHandler(registry))
	server.AddHealthHandler(router, db)
	router.HandleFunc("/migrate", func(resp http.ResponseWriter, _ *http.Request) {
		err := runner.StartMigration()
		if err != nil {
			resp.Write([]byte(fmt.Sprintf("%v", err.Error())))
			logger.Info(err.Error())
		}
		resp.WriteHeader(http.StatusOK)
	})

	port := viper.GetString(apiPortProp)
	host := viper.GetString(apiHostProp)
	logger.Info("Serving", lager.Data{"port": port})
	http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), router)
}

func labelName(label string) string {
	switch label {
	case "":
		return "none"
	default:
		return label
	}
}
