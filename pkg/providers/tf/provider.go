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

package tf

import (
	"context"
	"fmt"
	"time"

	"code.cloudfoundry.org/lager/v3"
	"github.com/hashicorp/go-version"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/executor"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/invoker"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/varcontext"
)

const (
	InProgress = "in progress"
	Succeeded  = "succeeded"
	Failed     = "failed"
)

// NewTerraformProvider creates a new ServiceProvider backed by Terraform module definitions for provision and bind.
func NewTerraformProvider(
	tfBinContext executor.TFBinariesContext,
	invokerBuilder invoker.TerraformInvokerBuilder,
	logger lager.Logger,
	serviceDefinition TfServiceDefinitionV1,
	deploymentManager DeploymentManagerInterface,
) *TerraformProvider {
	return &TerraformProvider{
		tfBinContext:               tfBinContext,
		TerraformInvokerBuilder:    invokerBuilder,
		serviceDefinition:          serviceDefinition,
		logger:                     logger.Session("terraform-" + serviceDefinition.Name),
		DeploymentManagerInterface: deploymentManager,
	}
}

type TerraformProvider struct {
	tfBinContext executor.TFBinariesContext
	invoker.TerraformInvokerBuilder
	logger            lager.Logger
	serviceDefinition TfServiceDefinitionV1
	DeploymentManagerInterface
}

func (provider *TerraformProvider) DefaultInvoker() invoker.TerraformInvoker {
	return provider.VersionedInvoker(provider.tfBinContext.DefaultTfVersion)
}

func (provider *TerraformProvider) VersionedInvoker(version *version.Version) invoker.TerraformInvoker {
	return provider.VersionedTerraformInvoker(version)
}

func (provider *TerraformProvider) create(ctx context.Context, vars *varcontext.VarContext, action TfServiceDefinitionV1Action, operationType string) error {
	tfID := vars.GetString("tf_id")
	if err := vars.Error(); err != nil {
		return err
	}

	newWorkspace, err := workspace.NewWorkspace(vars.ToMap(), action.Template, action.Templates, []workspace.ParameterMapping{}, []string{}, []workspace.ParameterMapping{})
	if err != nil {
		return fmt.Errorf("error creating workspace: %w", err)
	}

	deployment, err := provider.CreateAndSaveDeployment(tfID, newWorkspace)
	if err != nil {
		provider.logger.Error("deployment create failed", err)
		return fmt.Errorf("deployment create failed: %w", err)
	}

	if err := provider.MarkOperationStarted(&deployment, operationType); err != nil {
		return fmt.Errorf("error marking job started: %w", err)
	}

	go func() {
		var err error
		if vars.HasKey("vacant") && vars.GetBool("vacant") {
			newWorkspace.State = []byte(`{"version":4}`) // Minimum state required for anything to work
		} else {
			err = provider.DefaultInvoker().Apply(ctx, newWorkspace)
		}
		err = provider.MarkOperationFinished(&deployment, err)
		if err != nil {
			provider.logger.Error("MarkOperationFinished", err)
		}
	}()

	return nil
}

func (provider *TerraformProvider) destroy(ctx context.Context, deploymentID string, templateVars map[string]any, operationType string) error {
	deployment, err := provider.GetTerraformDeployment(deploymentID)
	if err != nil {
		return err
	}

	tfWorkspace := deployment.TFWorkspace()

	if err := tfWorkspace.RemovePreventDestroy(); err != nil {
		return err
	}

	inputList, err := tfWorkspace.Modules[0].Inputs()
	if err != nil {
		return err
	}

	limitedConfig := make(map[string]any)
	for _, name := range inputList {
		limitedConfig[name] = templateVars[name]
	}

	tfWorkspace.Instances[0].Configuration = limitedConfig

	fmt.Println("provider.MarkOperationStarted")
	if err := provider.MarkOperationStarted(&deployment, operationType); err != nil {
		return err
	}

	go func() {
		err = provider.DefaultInvoker().Destroy(ctx, tfWorkspace)
		_ = provider.MarkOperationFinished(&deployment, err)
	}()

	return nil
}

func (provider *TerraformProvider) Wait(ctx context.Context, id string) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		case <-time.After(1 * time.Second):
			isDone, _, _, err := provider.OperationStatus(id)
			if isDone {
				return err
			}
		}
	}
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . DeploymentManagerInterface
type DeploymentManagerInterface interface {
	GetTerraformDeployment(deploymentID string) (storage.TerraformDeployment, error)
	CreateAndSaveDeployment(deploymentID string, workspace *workspace.TerraformWorkspace) (storage.TerraformDeployment, error)
	MarkOperationStarted(deployment *storage.TerraformDeployment, operationType string) error
	MarkOperationFinished(deployment *storage.TerraformDeployment, err error) error
	OperationStatus(deploymentID string) (bool, string, string, error)
	UpdateWorkspaceHCL(deploymentID string, serviceDefinitionAction TfServiceDefinitionV1Action, templateVars map[string]any) error
	GetBindingDeployments(deploymentID string) ([]storage.TerraformDeployment, error)
	DeleteTerraformDeployment(deploymentID string) error
	ResetOperationType(deploymentID string) error
}
