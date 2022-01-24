package migrator

import (
	"context"
	"fmt"
	"os"
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
	tmpDir    string
}

func (r *MigrationRunner) StartMigration() error {
	r.logger.Info("Starting Migration")

	instances, err := r.storage.GetServiceInstances()
	if err != nil {
		return err
	}
	start := time.Now()
	failedUpdates := 0

	for _, instance := range instances {

		jobRunner := tf.NewTfJobRunner(nil, r.storage)

		r.logger.Info("Creating executor for 0.13")

		jobRunner.Executor = r.ExecutorFor013()
		err := jobRunner.MigrateTo013(context.TODO(), "tf:"+instance.GUID+":")
		if err != nil {
			failedUpdates += 1
			r.logger.Error(fmt.Sprintf("Failed to update instance: %v to TF 0.13.", instance.GUID), err)
			continue
		}
		err = r.Cleanup()
		if err != nil {
			r.logger.Error("Cleanup error", err)
			return err
		}

		r.logger.Info("Creating executor for 0.14")
		jobRunner.Executor = r.ExecutorFor014()
		err = jobRunner.MigrateTo014(context.TODO(), "tf:"+instance.GUID+":")
		if err != nil {
			failedUpdates += 1
			r.logger.Error(fmt.Sprintf("Failed to update instance: %v to TF 0.14.", instance.GUID), err)
			continue
		}
		err = r.Cleanup()
		if err != nil {
			r.logger.Error("Cleanup error", err)
			return err
		}

		r.logger.Info("Creating executor for 1.0")
		jobRunner.Executor = r.ExecutorFor10()
		err = jobRunner.MigrateTo10(context.TODO(), "tf:"+instance.GUID+":")
		if err != nil {
			failedUpdates += 1
			r.logger.Error(fmt.Sprintf("Failed to update instance: %v to TF 1.0.", instance.GUID), err)
			continue
		}
		err = r.Cleanup()
		if err != nil {
			r.logger.Error("Cleanup error", err)
			return err
		}

		r.logger.Info("Creating executor for 1.1")
		jobRunner.Executor = r.ExecutorFor11()
		err = jobRunner.MigrateTo11(context.TODO(), "tf:"+instance.GUID+":")
		if err != nil {
			failedUpdates += 1
			r.logger.Error(fmt.Sprintf("Failed to update instance: %v to TF 1.1.", instance.GUID), err)
			continue
		}
		err = r.Cleanup()
		if err != nil {
			r.logger.Error("Cleanup error", err)
			return err
		}

	}
	r.logger.Info(fmt.Sprintf("Number of instances: %d\n", len(instances)))
	r.logger.Info(fmt.Sprintf("Total Failures: %d\n", failedUpdates))
	r.logger.Info(fmt.Sprintf("Total Runtime: %f\n", time.Since(start).Minutes()))
	r.logger.Info(fmt.Sprintf("Avg runtime per instance: %f\n", time.Since(start).Seconds()/float64(len(instances))))

	return nil
}

func (r *MigrationRunner) ExecutorFor013() wrapper.TerraformExecutor {
	dir, _ := os.MkdirTemp("", "brokerpak")
	r.tmpDir = dir

	executor, err := r.registrar.CreateTerraformExecutor(dir, "0.13.7")
	if err != nil {
		panic(err)
	}
	return executor
}

func (r *MigrationRunner) ExecutorFor014() wrapper.TerraformExecutor {
	dir, _ := os.MkdirTemp("", "brokerpak")
	r.tmpDir = dir

	executor, err := r.registrar.CreateTerraformExecutor(dir, "0.14.9")
	if err != nil {
		panic(err)
	}
	return executor
}

func (r *MigrationRunner) ExecutorFor10() wrapper.TerraformExecutor {
	dir, _ := os.MkdirTemp("", "brokerpak")
	r.tmpDir = dir

	executor, err := r.registrar.CreateTerraformExecutor(dir, "1.0.9")
	if err != nil {
		panic(err)
	}
	return executor
}

func (r *MigrationRunner) ExecutorFor11() wrapper.TerraformExecutor {
	dir, _ := os.MkdirTemp("", "brokerpak")
	r.tmpDir = dir

	executor, err := r.registrar.CreateTerraformExecutor(dir, "1.1.3")
	if err != nil {
		panic(err)
	}
	return executor
}

func (r *MigrationRunner) Cleanup() error {
	r.logger.Info(fmt.Sprintf("Removing dir: %s", r.tmpDir))
	return os.RemoveAll(r.tmpDir)
}

func New(config *brokers.BrokerConfig, logger lager.Logger, storage *storage.Storage, registrar *brokerpak.Registrar) (*MigrationRunner, error) {
	return &MigrationRunner{
		config:    config,
		logger:    logger,
		storage:   storage,
		registrar: registrar,
	}, nil

}
