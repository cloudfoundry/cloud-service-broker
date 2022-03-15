package wrapper

import "fmt"

type TerraformCommand interface {
	Command() []string
}

func NewInit012Command(pluginDir string) TerraformCommand {
	return init012Command{pluginDir: pluginDir}
}

type init012Command struct {
	pluginDir string
}

func (cmd init012Command) Command() []string {
	return []string{"init", fmt.Sprintf("-plugin-dir=%s", cmd.pluginDir), "-get-plugins=false", "-no-color"}
}

func NewInitCommand(pluginDir string) TerraformCommand {
	return initCommand{pluginDir: pluginDir}
}

type initCommand struct {
	pluginDir string
}

func (cmd initCommand) Command() []string {
	return []string{"init", fmt.Sprintf("-plugin-dir=%s", cmd.pluginDir), "-no-color"}
}

type ApplyCommand struct{}

func (ApplyCommand) Command() []string {
	return []string{"apply", "-auto-approve", "-no-color"}
}

type DestroyCommand struct{}

func (DestroyCommand) Command() []string {
	return []string{"destroy", "-auto-approve", "-no-color"}
}

type ShowCommand struct{}

func (ShowCommand) Command() []string {
	return []string{"show", "-no-color"}
}

type PlanCommand struct{}

func (PlanCommand) Command() []string {
	return []string{"plan", "-no-color"}
}

type ImportCommand struct {
	Addr string
	ID   string
}

func (cmd ImportCommand) Command() []string {
	return []string{"import", cmd.Addr, cmd.ID}
}

type renameProviderCommand struct {
	oldProviderName string
	newProviderName string
}

func (cmd renameProviderCommand) Command() []string {
	return []string{"state", "replace-provider", cmd.oldProviderName, cmd.newProviderName}
}

func NewRenameProviderCommand(oldProviderName, newProviderName string) TerraformCommand {
	return renameProviderCommand{oldProviderName: oldProviderName, newProviderName: newProviderName}
}
