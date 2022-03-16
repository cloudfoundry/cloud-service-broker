package command

import (
	"fmt"
)

func NewInit012Command(pluginDir string) TerraformCommand {
	return init012{pluginDir: pluginDir}
}

type init012 struct {
	pluginDir string
}

func (cmd init012) Command() []string {
	return []string{"init", fmt.Sprintf("-plugin-dir=%s", cmd.pluginDir), "-get-plugins=false", "-no-color"}
}

func NewInit(pluginDir string) TerraformCommand {
	return initCommand{pluginDir: pluginDir}
}

type initCommand struct {
	pluginDir string
}

func (cmd initCommand) Command() []string {
	return []string{"init", fmt.Sprintf("-plugin-dir=%s", cmd.pluginDir), "-no-color"}
}

type Apply struct{}

func (Apply) Command() []string {
	return []string{"apply", "-auto-approve", "-no-color"}
}

type Destroy struct{}

func (Destroy) Command() []string {
	return []string{"destroy", "-auto-approve", "-no-color"}
}

type Show struct{}

func (Show) Command() []string {
	return []string{"show", "-no-color"}
}

type Plan struct{}

func (Plan) Command() []string {
	return []string{"plan", "-no-color"}
}

type Import struct {
	Addr string
	ID   string
}

func (cmd Import) Command() []string {
	return []string{"import", cmd.Addr, cmd.ID}
}

type renameProvider struct {
	oldProviderName string
	newProviderName string
}

func (cmd renameProvider) Command() []string {
	return []string{"state", "replace-provider", "-auto-approve", cmd.oldProviderName, cmd.newProviderName}
}

func NewRenameProvider(oldProviderName, newProviderName string) TerraformCommand {
	return renameProvider{oldProviderName: oldProviderName, newProviderName: newProviderName}
}
