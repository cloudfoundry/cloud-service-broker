package tf

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . WorkspaceBuilder

type WorkspaceBuilder interface {
	CreateWorkspace(deployment storage.TerraformDeployment) (Workspace, error)
}

func NewWorkspaceFactory() WorkspaceFactory {
	return WorkspaceFactory{}
}

type WorkspaceFactory struct {
}

func (w WorkspaceFactory) CreateWorkspace(deployment storage.TerraformDeployment) (Workspace, error) {
	return wrapper.DeserializeWorkspace(deployment.Workspace)
}
