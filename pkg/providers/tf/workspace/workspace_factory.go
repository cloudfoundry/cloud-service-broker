package workspace

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
)

func NewWorkspaceFactory() WorkspaceFactory {
	return WorkspaceFactory{}
}

type WorkspaceFactory struct {
}

func (w WorkspaceFactory) CreateWorkspace(deployment storage.TerraformDeployment) (Workspace, error) {
	return DeserializeWorkspace(deployment.Workspace)
}
