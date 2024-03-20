package command

import (
	"fmt"
)

func NewInit(pluginDir string) TerraformCommand {
	return initCommand{pluginDir: pluginDir}
}

type initCommand struct {
	pluginDir string
}

func (cmd initCommand) Command() []string {
	return []string{"init", fmt.Sprintf("-plugin-dir=%s", cmd.pluginDir), "-no-color"}
}

type apply struct{}

func NewApply() TerraformCommand {
	return apply{}
}
func (apply) Command() []string {
	return []string{"apply", "-auto-approve", "-no-color"}
}

func NewDestroy() TerraformCommand {
	return destroy{}
}

type destroy struct{}

func (destroy) Command() []string {
	return []string{"destroy", "-auto-approve", "-no-color"}
}

func NewShow() TerraformCommand {
	return show{}
}

type show struct{}

func (show) Command() []string {
	return []string{"show", "-no-color"}
}

func NewPlan() TerraformCommand {
	return plan{}
}

type plan struct{}

func (plan) Command() []string {
	return []string{"plan", "-no-color"}
}

func NewImport(addr, id string) TerraformCommand {
	return importCmd{Addr: addr, ID: id}
}

type importCmd struct {
	Addr string
	ID   string
}

func (cmd importCmd) Command() []string {
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
