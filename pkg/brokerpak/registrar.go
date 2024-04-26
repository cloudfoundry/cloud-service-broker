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

package brokerpak

import (
	"encoding/json"
	"fmt"
	"os"

	"code.cloudfoundry.org/lager/v3"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/brokerpak/manifest"
	"github.com/cloudfoundry/cloud-service-broker/v3/internal/brokerpak/reader"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/featureflags"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/varcontext"
	"github.com/cloudfoundry/cloud-service-broker/v3/utils"
)

type registrarWalkFunc func(name string, pak BrokerpakSourceConfig, vc *varcontext.VarContext) error

// Registrar is responsible for registering brokerpaks with BrokerRegistries
// subject to the settings provided by a ServerConfig like injecting
// environment variables and skipping certain services.
type Registrar struct {
	config *ServerConfig
}

// Register fetches the brokerpaks and registers them with the given registry.
func (r *Registrar) Register(registry broker.BrokerRegistry) error {
	registerLogger := utils.NewLogger("brokerpak-registration")

	return r.walk(func(name string, pak BrokerpakSourceConfig, vc *varcontext.VarContext) error {
		registerLogger.Info("registering", lager.Data{
			"name":              name,
			"location":          pak.BrokerpakURI,
			"notes":             pak.Notes,
			"excluded-services": pak.ExcludedServicesSlice(),
			"prefix":            pak.ServicePrefix,
		})

		brokerPak, err := reader.DownloadAndOpenBrokerpak(pak.BrokerpakURI)
		if err != nil {
			return fmt.Errorf("couldn't open brokerpak: %q: %v", pak.BrokerpakURI, err)
		}
		defer brokerPak.Close()

		tfBinariesContext, err := r.extractTfBinaries(brokerPak, vc)
		if err != nil {
			return err
		}

		// register the services
		services, err := brokerPak.Services()
		if err != nil {
			return err
		}

		var maintenanceInfo *domain.MaintenanceInfo
		if featureflags.Enabled(featureflags.TfUpgradeEnabled) {
			maintenanceInfo = &domain.MaintenanceInfo{
				Version:     tfBinariesContext.DefaultTfVersion.String(),
				Description: fmt.Sprintf(`This upgrade provides support for OpenTofu version: %s. The upgrade operation will take a while. The instance and all associated bindings will be upgraded.`, tfBinariesContext.DefaultTfVersion.String()),
			}
		}

		defns, err := r.toDefinitions(services, pak, tfBinariesContext, maintenanceInfo)
		if err != nil {
			return err
		}

		for _, defn := range defns {
			err := registry.Register(defn, maintenanceInfo)
			if err != nil {
				return err
			}
		}

		errs := registry.Validate()
		if errs != nil {
			return errs
		}

		mf, err := brokerPak.Manifest()
		if err != nil {
			return fmt.Errorf("error reading brokerpak manifest: %w", err)
		}

		for env, config := range mf.EnvConfigMapping {
			viper.BindEnv(config, env)
		}

		registerLogger.Info("registration-successful", lager.Data{"version": mf.Version})
		return nil
	})
}

func (r *Registrar) toDefinitions(services []tf.TfServiceDefinitionV1, config BrokerpakSourceConfig, tfBinariesContext executor.TFBinariesContext, maintenanceInfo *domain.MaintenanceInfo) ([]*broker.ServiceDefinition, error) {
	var out []*broker.ServiceDefinition
	globalLabels, err := r.config.GetGlobalLabels()
	if err != nil {
		return nil, err
	}

	toIgnore := utils.NewStringSet(config.ExcludedServicesSlice()...)
	for _, svc := range services {
		if toIgnore.Contains(svc.ID) {
			continue
		}

		svc.Name = config.ServicePrefix + svc.Name

		bs, err := svc.ToService(tfBinariesContext, maintenanceInfo)
		if err != nil {
			return nil, err
		}

		bs.GlobalLabels = globalLabels

		out = append(out, bs)
	}

	return out, nil
}

func (r *Registrar) extractTfBinaries(brokerPak *reader.BrokerPakReader, vc *varcontext.VarContext) (executor.TFBinariesContext, error) {
	dir, err := os.MkdirTemp("", "brokerpak")
	if err != nil {
		return executor.TFBinariesContext{}, err
	}

	// extract the Terraform directory
	if err := brokerPak.ExtractPlatformBins(dir); err != nil {
		return executor.TFBinariesContext{}, err
	}

	manifest, err := brokerPak.Manifest()
	if err != nil {
		return executor.TFBinariesContext{}, err
	}

	tfVersion, err := manifest.DefaultTerraformVersion()
	if err != nil {
		return executor.TFBinariesContext{}, err
	}

	return executor.TFBinariesContext{
		Dir:                  dir,
		DefaultTfVersion:     tfVersion,
		Params:               resolveParameters(manifest.Parameters, vc),
		TfUpgradePath:        manifest.TerraformUpgradePath,
		ProviderReplacements: manifest.TerraformStateProviderReplacements,
	}, nil
}

func (r *Registrar) walk(callback registrarWalkFunc) error {
	for name, pak := range r.config.Brokerpaks {
		vc, err := varcontext.Builder().
			MergeJSONObject(json.RawMessage(r.config.Config)).
			MergeJSONObject(json.RawMessage(pak.Config)).
			Build()

		if err != nil {
			return fmt.Errorf("couldn't merge config for brokerpak %q: %v", name, err)
		}

		if err := callback(name, pak, vc); err != nil {
			return err
		}
	}

	return nil
}

// NewRegistrar constructs a new registrar with the given configuration.
// Registrar expects to become the owner of the configuration afterwards.
func NewRegistrar(sc *ServerConfig) *Registrar {
	return &Registrar{config: sc}
}

// resolveParameters resolves environment variables from the given global and
// brokerpak specific.
func resolveParameters(params []manifest.Parameter, vc *varcontext.VarContext) map[string]string {
	out := make(map[string]string)

	context := vc.ToMap()
	for _, p := range params {
		val, ok := context[p.Name]
		if ok {
			out[p.Name] = cast.ToString(val)
		}
	}

	return out
}
