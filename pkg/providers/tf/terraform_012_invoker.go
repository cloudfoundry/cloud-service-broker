package tf

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
)

func NewTerraform012Invoker(executor wrapper.TerraformExecutor, pluginDirectory string) TerraformInvoker {
	return Terraform012Invoker{executor: executor, pluginDirectory: pluginDirectory}
}

type Terraform012Invoker struct {
	executor        wrapper.TerraformExecutor
	pluginDirectory string
}

func (cmd Terraform012Invoker) Apply(ctx context.Context, workspace Workspace) error {
	_, err := workspace.Execute(ctx, cmd.executor,
		wrapper.NewInit012Command(cmd.pluginDirectory),
		wrapper.ApplyCommand{})
	return err
}

func (cmd Terraform012Invoker) Show(ctx context.Context, workspace Workspace) (string, error) {
	output, err := workspace.Execute(ctx, cmd.executor,
		wrapper.NewInit012Command(cmd.pluginDirectory),
		wrapper.ShowCommand{})
	return output.StdOut, err
}

func (cmd Terraform012Invoker) Destroy(ctx context.Context, workspace Workspace) error {
	_, err := workspace.Execute(ctx, cmd.executor,
		wrapper.NewInit012Command(cmd.pluginDirectory),
		wrapper.DestroyCommand{})
	return err
}

func (cmd Terraform012Invoker) Plan(ctx context.Context, workspace Workspace) (wrapper.ExecutionOutput, error) {
	return workspace.Execute(ctx, cmd.executor,
		wrapper.NewInit012Command(cmd.pluginDirectory),
		wrapper.PlanCommand{})
}

func (cmd Terraform012Invoker) Import(ctx context.Context, workspace Workspace, resources map[string]string) error {
	commands := []wrapper.TerraformCommand{
		wrapper.NewInit012Command(cmd.pluginDirectory),
	}
	for resource, id := range resources {
		commands = append(commands, wrapper.ImportCommand{Addr: resource, ID: id})
	}

	_, err := workspace.Execute(ctx, cmd.executor, commands...)
	return err
}
