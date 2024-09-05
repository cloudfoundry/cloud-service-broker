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
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"code.cloudfoundry.org/lager/v3"
	osbapiBroker "github.com/cloudfoundry/cloud-service-broker/v2/brokerapi/broker"
	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/displaycatalog"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/encryption"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/infohandler"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	pakBroker "github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/brokerpak"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/server"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/toggles"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils"
	"github.com/google/uuid"
	"github.com/pivotal-cf/brokerapi/v11"
	"github.com/pivotal-cf/brokerapi/v11/auth"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const (
	apiUserProp         = "api.user"
	apiPasswordProp     = "api.password"
	apiPortProp         = "api.port"
	apiHostProp         = "api.host"
	tlsCertCaBundleProp = "api.certCaBundle"
	tlsKeyProp          = "api.tlsKey"
	encryptionPasswords = "db.encryption.passwords"
	encryptionEnabled   = "db.encryption.enabled"

	shutdownTimeout = time.Hour
)

var cfCompatibilityToggle = toggles.Features.Toggle("enable-cf-sharing", false, `Set all services to have the Sharable flag so they can be shared
	across spaces in PCF.`)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:     "serve",
		GroupID: "broker",
		Short:   "Start the service broker",
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

	_ = viper.BindEnv(apiUserProp, "SECURITY_USER_NAME")
	_ = viper.BindEnv(apiPasswordProp, "SECURITY_USER_PASSWORD")
	_ = viper.BindEnv(apiPortProp, "PORT")
	_ = viper.BindEnv(apiHostProp, "CSB_LISTENER_HOST")
	_ = viper.BindEnv(encryptionPasswords, "ENCRYPTION_PASSWORDS")
	_ = viper.BindEnv(encryptionEnabled, "ENCRYPTION_ENABLED")
	_ = viper.BindEnv(tlsCertCaBundleProp, "TLS_CERT_CHAIN")
	_ = viper.BindEnv(tlsKeyProp, "TLS_PRIVATE_KEY")
}

func serve() {
	logger := utils.NewLogger("cloud-service-broker")
	logger.Info("starting", lager.Data{"version": utils.Version})
	db := dbservice.NewWithMigrations(logger)
	encryptor := setupDBEncryption(db, logger)

	// init broker
	cfg, err := osbapiBroker.NewBrokerConfigFromEnv(logger)
	if err != nil {
		logger.Fatal("Error initializing service broker config", err)
	}
	var serviceBroker domain.ServiceBroker
	csbStore := storage.New(db, encryptor)
	serviceBroker, err = osbapiBroker.New(cfg, csbStore, logger)
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
	switch {
	case len(services) == 0:
		logger.Fatal("refusing to start", fmt.Errorf("no services are defined"))
	case err != nil:
		logger.Error("creating service catalog", err)
	default:
		logger.Info("service catalog", lager.Data{"catalog": displaycatalog.DisplayCatalog(services)})
	}

	brokerAPI := brokerapi.New(serviceBroker, slog.New(lager.NewHandler(logger)), credentials)

	sqldb, err := db.DB()
	if err != nil {
		logger.Error("failed to get database connection", err)
	}
	startServer(cfg.Registry, sqldb, brokerAPI, storage.New(db, encryptor), credentials)
}

func serveDocs() {
	logger := utils.NewLogger("cloud-service-broker")
	// init broker
	registry := pakBroker.BrokerRegistry{}
	if err := brokerpak.RegisterAll(registry); err != nil {
		logger.Error("loading brokerpaks", err)
	}

	startServer(registry, nil, nil, nil, brokerapi.BrokerCredentials{})
}

func setupDBEncryption(db *gorm.DB, logger lager.Logger) storage.Encryptor {
	config, err := encryption.ParseConfiguration(db, viper.GetBool(encryptionEnabled), viper.Get(encryptionPasswords))
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

func startServer(registry pakBroker.BrokerRegistry, db *sql.DB, brokerapi http.Handler, store *storage.Storage, credentials brokerapi.BrokerCredentials) {
	logger := utils.NewLogger("cloud-service-broker")

	docsHandler := server.DocsHandler(registry)

	router := http.NewServeMux()
	router.Handle("/docs", docsHandler)
	router.HandleFunc("/examples", server.NewExampleHandler(registry))
	server.AddHealthHandler(router, db)
	router.HandleFunc("/info", infohandler.NewDefault())
	router.Handle("/import_state/{guid}", auth.NewWrapper(credentials.Username, credentials.Password).Wrap(importStateHandler(store)))

	router.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		switch {
		case req.URL.Path == "/":
			docsHandler.ServeHTTP(res, req)
		case strings.HasPrefix(req.URL.Path, "/v2") && brokerapi != nil:
			brokerapi.ServeHTTP(res, req)
		default:
			http.NotFound(res, req)
		}
	})

	port := viper.GetString(apiPortProp)
	host := viper.GetString(apiHostProp)
	logger.Info("Serving", lager.Data{"port": port})

	tlsCertCaBundle := viper.GetString(tlsCertCaBundleProp)
	tlsKey := viper.GetString(tlsKeyProp)

	logger.Info("tlsCertCaBundle", lager.Data{"tlsCertCaBundle": tlsCertCaBundle})
	logger.Info("tlsKey", lager.Data{"tlsKey": tlsKey})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: router,
	}

	go func() {
		var err error
		if tlsCertCaBundle != "" && tlsKey != "" {
			err = httpServer.ListenAndServeTLS(tlsCertCaBundle, tlsKey)
		} else {
			err = httpServer.ListenAndServe()
		}
		if err == http.ErrServerClosed {
			logger.Info("shutting down csb")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGKILL, syscall.SIGTERM)

	signalReceived := <-sigChan

	switch signalReceived {

	case syscall.SIGTERM:
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), shutdownTimeout)
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Fatal("shutdown error: %v", err)
		}
		logger.Info("received SIGTERM, server is shutting down gracefully allowing for in flight work to finish")
		defer shutdownRelease()
		for store.LockFilesExist() {
			logger.Info("draining csb instance")
			time.Sleep(time.Second * 1)
		}
		logger.Info("draining complete")
		logger.Info("shutdown complete")
	case syscall.SIGKILL:
		logger.Info("received SIGKILL, server is shutting down immediately. In flight operations will not finish and their state is potentially lost.")
		if err := httpServer.Close(); err != nil {
			logger.Error("shutdown", err)
		}
	default:
		logger.Info("csb does not handle the interrupt signal")
	}
}

func labelName(label string) string {
	switch label {
	case "":
		return "none"
	default:
		return label
	}
}

func importStateHandler(store *storage.Storage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		guid := r.PathValue("guid")
		if guid == "" {
			http.Error(w, "GUID is required", http.StatusBadRequest)
			return
		}
		if err := uuid.Validate(guid); err != nil {
			http.Error(w, "not a valid GUID", http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read body: %s", err), http.StatusBadRequest)
			return
		}
		var b any
		if err := json.Unmarshal(body, &b); err != nil {
			http.Error(w, fmt.Sprintf("problem parsing body as JSON: %s", err), http.StatusBadRequest)
			return
		}

		tfID := fmt.Sprintf("tf:%s:", guid)
		deployment, err := store.GetTerraformDeployment(tfID)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not find TF ID: %s", tfID), http.StatusNotFound)
			return
		}

		tfw, ok := deployment.Workspace.(*workspace.TerraformWorkspace)
		if !ok {
			http.Error(w, "failed cast to *workspace.TerraformWorkspace", http.StatusInternalServerError)
			return
		}

		tfw.State = body
		if err := store.StoreTerraformDeployment(deployment); err != nil {
			http.Error(w, fmt.Sprintf("failed to store deployment: %s", err), http.StatusInternalServerError)
			return
		}
	})
}
