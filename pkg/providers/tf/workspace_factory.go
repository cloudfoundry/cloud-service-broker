package tf

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . WorkspaceFactory

type WorkspaceFactory interface {
	CreateWorkspace(deployment storage.TerraformDeployment) (Workspace, error)
}

func NewWorkspaceFactoryImpl() WorkspaceFactoryImpl {
	return WorkspaceFactoryImpl{}
}

type WorkspaceFactoryImpl struct {
}

func (w WorkspaceFactoryImpl) CreateWorkspace(deployment storage.TerraformDeployment) (Workspace, error) {
	return wrapper.DeserializeWorkspace(deployment.Workspace)
}
