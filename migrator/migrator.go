package migrator

import (
	"context"
	"fmt"
	"time"

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
	r.logger.Info("Starting Migration")

	instances, err := r.storage.GetServiceInstances()
	if err != nil {
		return err
	}
	start := time.Now()
	for _, instance := range instances {

		jobRunner := tf.NewTfJobRunner(nil, r.storage)

		jobRunner.Executor = r.ExecutorFor013()
		err := jobRunner.MigrateTo013(context.TODO(), "tf:"+instance.GUID+":")
		if err != nil {
			return err
		}

		jobRunner.Executor = r.ExecutorFor014()
		err = jobRunner.MigrateTo014(context.TODO(), "tf:"+instance.GUID+":")
		if err != nil {
			return err
		}

		jobRunner.Executor = r.ExecutorFor10()
		err = jobRunner.MigrateTo10(context.TODO(), "tf:"+instance.GUID+":")
		if err != nil {
			return err
		}

		jobRunner.Executor = r.ExecutorFor11()
		err = jobRunner.MigrateTo11(context.TODO(), "tf:"+instance.GUID+":")
		if err != nil {
			return err
		}

	}
	fmt.Printf("Number of instances: %v\n", len(instances))
	fmt.Printf("Total Runtime: %v\n", time.Since(start))
	fmt.Printf("Avg runtime per instance: %v\n", time.Since(start).Seconds()/float64(len(instances)))
	return nil
}

func (r *MigrationRunner) ExecutorFor013() wrapper.TerraformExecutor {
	executor, err := r.registrar.CreateTerraformExecutor("0.13.7")
	if err != nil {
		panic(err)
	}
	return executor
}

func (r *MigrationRunner) ExecutorFor014() wrapper.TerraformExecutor {
	executor, err := r.registrar.CreateTerraformExecutor("0.14.9")
	if err != nil {
		panic(err)
	}
	return executor
}

func (r *MigrationRunner) ExecutorFor10() wrapper.TerraformExecutor {
	executor, err := r.registrar.CreateTerraformExecutor("1.0.9")
	if err != nil {
		panic(err)
	}
	return executor
}

func (r *MigrationRunner) ExecutorFor11() wrapper.TerraformExecutor {
	executor, err := r.registrar.CreateTerraformExecutor("1.1.3")
	if err != nil {
		panic(err)
	}
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
