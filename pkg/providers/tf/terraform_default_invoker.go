package tf

import (
	"context"

	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf/wrapper"
)

func NewTerraformDefaultInvoker(executor wrapper.TerraformExecutor, pluginDirectory string, pluginRenames map[string]string) TerraformInvoker {
	return TerraformDefaultInvoker{executor: executor, pluginDirectory: pluginDirectory, providerReplaceGenerator: pluginRenames}
}

type TerraformDefaultInvoker struct {
	executor        wrapper.TerraformExecutor
	pluginDirectory string
	providerReplaceGenerator
}

func (cmd TerraformDefaultInvoker) Apply(ctx context.Context, workspace Workspace) error {
	var commands []wrapper.TerraformCommand
	if workspace.HasState() {
		commands = cmd.ReplacementCommands()
	}
	commands = append(commands, wrapper.NewInitCommand(cmd.pluginDirectory), wrapper.ApplyCommand{})

	_, err := workspace.Execute(ctx, cmd.executor, commands...)
	return err
}

func (cmd TerraformDefaultInvoker) Show(ctx context.Context, workspace Workspace) (string, error) {
	output, err := workspace.Execute(ctx, cmd.executor,
		wrapper.NewInitCommand(cmd.pluginDirectory),
		wrapper.ShowCommand{})
	return output.StdOut, err
}

func (cmd TerraformDefaultInvoker) Destroy(ctx context.Context, workspace Workspace) error {
	_, err := workspace.Execute(ctx, cmd.executor,
		wrapper.NewInitCommand(cmd.pluginDirectory),
		wrapper.DestroyCommand{})
	return err
}

func (cmd TerraformDefaultInvoker) Plan(ctx context.Context, workspace Workspace) (wrapper.ExecutionOutput, error) {
	return workspace.Execute(ctx, cmd.executor,
		wrapper.NewInitCommand(cmd.pluginDirectory),
		wrapper.PlanCommand{})
}

func (cmd TerraformDefaultInvoker) Import(ctx context.Context, workspace Workspace, resources map[string]string) error {
	commands := []wrapper.TerraformCommand{
		wrapper.NewInitCommand(cmd.pluginDirectory),
	}
	for resource, id := range resources {
		commands = append(commands, wrapper.ImportCommand{Addr: resource, ID: id})
	}

	_, err := workspace.Execute(ctx, cmd.executor, commands...)
	return err
}

type providerReplaceGenerator map[string]string

func (replace providerReplaceGenerator) ReplacementCommands() []wrapper.TerraformCommand {
	var commands []wrapper.TerraformCommand

	for old, new := range replace {
		commands = append(commands, wrapper.NewRenameProviderCommand(old, new))
	}

	return commands
}
