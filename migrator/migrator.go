package migrator

import (
	"context"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/brokerpak"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/cloud-service-broker/brokerapi/brokers"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf/wrapper"
)

type MigrationRunner struct {
	storage   *storage.Storage
	config    *brokers.BrokerConfig
	logger    lager.Logger
	registrar *brokerpak.Registrar
}

func (r *MigrationRunner) StartMigration() error {
	// Get Terraform binaries
	// Unpack Terraform binaries
	// Unpack Terraform Providers
	// Create FS from terraform_deployments DB entry
	// For each terraform version Run terraform init/update/apply
	// Save FS back to DB

	// get default runner(used by broker)
	//

	r.logger.Info("Starting Migration")

	instances, err := r.storage.GetServiceInstances()
	if err != nil {
		return err
	}

	for _, instance := range instances {
		jobRunner := tf.NewTfJobRunner(nil, r.storage)

		jobRunner.Executor = r.ExecutorFor012()
		err := jobRunner.MigrateTo013(context.TODO(), "tf:"+instance.GUID+":")
		if err != nil {
			return err
		}

		//jobRunner.Executor = ExecutorFor013()
		//jobRunner.MigrateTo014(context.TODO(), "tf:"+instance.GUID)
		//jobRunner.Executor = ExecutorFor014()
		//jobRunner.MigrateTo10(context.TODO(), "tf:"+instance.GUID)
		//jobRunner.Executor = ExecutorFor11()
		//jobRunner.Migrateto11(context.TODO(), "tf:"+instance.GUID)

		//
	}
	return nil
	// path for 0.12
	// path for 0.13
	// path for 0.14
	// path for 1.0
	// path for 1.1
}

func (r *MigrationRunner) ExecutorFor012() wrapper.TerraformExecutor {
	executor, _ := r.registrar.CreateTerraformExecutor("0.12.30")
	return executor
}

func New(config *brokers.BrokerConfig, logger lager.Logger, storage *storage.Storage, registrar *brokerpak.Registrar) (*MigrationRunner, error) {
	return &MigrationRunner{
		config:    config,
		logger:    logger,
		storage:   storage,
		registrar: registrar,
	}, nil

}
