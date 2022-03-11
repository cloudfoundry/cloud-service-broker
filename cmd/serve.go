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

	"code.cloudfoundry.org/lager"
	osbapiBroker "github.com/cloudfoundry/cloud-service-broker/brokerapi/broker"
	"github.com/cloudfoundry/cloud-service-broker/db_service"
	"github.com/cloudfoundry/cloud-service-broker/internal/encryption"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	pakBroker "github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/brokerpak"
	"github.com/cloudfoundry/cloud-service-broker/pkg/server"
	"github.com/cloudfoundry/cloud-service-broker/pkg/toggles"
	"github.com/cloudfoundry/cloud-service-broker/utils"
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
	logger.Info("starting", lager.Data{"version": utils.Version})
	db := db_service.New(logger)
	encryptor := setupDBEncryption(db, logger)

	// init broker
	cfg, err := osbapiBroker.NewBrokerConfigFromEnv(logger)
	if err != nil {
		logger.Fatal("Error initializing service broker config", err)
	}
	var serviceBroker domain.ServiceBroker
	serviceBroker, err = osbapiBroker.New(cfg, logger, storage.New(db, encryptor))
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
	startServer(cfg.Registry, sqldb, brokerAPI)
}

func serveDocs() {
	logger := utils.NewLogger("cloud-service-broker")
	// init broker
	registry := pakBroker.BrokerRegistry{}
	if err := brokerpak.RegisterAll(registry); err != nil {
		logger.Error("loading brokerpaks", err)
	}

	startServer(registry, nil, nil)
}

func setupDBEncryption(db *gorm.DB, logger lager.Logger) storage.Encryptor {
	config, err := encryption.ParseConfiguration(db, viper.GetBool(encryptionEnabled), viper.GetString(encryptionPasswords))
	if err != nil {
		logger.Fatal("Error parsing encryption configuration", err)
	}

	if config.Changed {
		if err := storage.New(db, config.RotationEncryptor).CheckAllRecords(); err != nil {
			logger.Fatal("refusing to encrypt the database as some fields cannot be successfully read", err)
		}

		logger.Info("rotating-database-encryption", lager.Data{"previous-primary": labelName(config.StoredPrimaryLabel), "new-primary": labelName(config.ConfiguredPrimaryLabel)})
		if err := storage.New(db, config.RotationEncryptor).UpdateAllRecords(); err != nil {
			logger.Fatal("Error rotating database encryption", err)
		}
		if err := encryption.UpdatePasswordMetadata(db, config.ConfiguredPrimaryLabel); err != nil {
			logger.Fatal("Error updating password metadata", err)
		}
	}

	err = storage.New(db, config.Encryptor).CheckAllRecords()
	switch {
	case err != nil:
		// This error denotes that there was a problem reading at least one database field.
		// If you see this error, examine the rows and the error message and try to correct the data if you can.
		// If there is data in the database that cannot be read, it may not be possible to update the service
		// instance or service binding that it relates to. This may not be a problem in the short term, but in
		// the longer term you should aim to delete the object. It may be necessary to raise an issue to get
		// assistance with this.
		logger.Error("database-field-error", err)
	case len(config.ToDeleteLabels) > 0:
		logger.Info("removing-state-password-metadata", lager.Data{"labels": config.ToDeleteLabels})
		if err := encryption.DeletePasswordMetadata(db, config.ToDeleteLabels); err != nil {
			logger.Fatal("Error deleting stale password metadata", err)
		}
	}

	logger.Info("database-encryption", lager.Data{"primary": labelName(config.ConfiguredPrimaryLabel)})
	return config.Encryptor
}

func startServer(registry pakBroker.BrokerRegistry, db *sql.DB, brokerapi http.Handler) {
	logger := utils.NewLogger("cloud-service-broker")

	router := mux.NewRouter()

	// match paths going to the brokerapi first
	if brokerapi != nil {
		router.PathPrefix("/v2").Handler(brokerapi)
	}

	server.AddDocsHandler(router, registry)
	router.HandleFunc("/examples", server.NewExampleHandler(registry))
	server.AddHealthHandler(router, db)

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
