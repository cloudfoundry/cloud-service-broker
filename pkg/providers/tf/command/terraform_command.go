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

func (cmd initCommand) Env() []string {
	return []string{}
}

type apply struct{}

func NewApply() TerraformCommand {
	return apply{}
}

func (cmd apply) Env() []string {
	return []string{}
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

func (cmd destroy) Env() []string {
	return []string{}
}

func NewShow() TerraformCommand {
	return show{}
}

type show struct{}

func (show) Command() []string {
	return []string{"show", "-no-color"}
}

func (cmd show) Env() []string {
	return []string{"OPENTOFU_STATEFILE_PROVIDER_ADDRESS_TRANSLATION=0"}
}

func NewPlan() TerraformCommand {
	return plan{}
}

type plan struct{}

func (plan) Command() []string {
	return []string{"plan", "-no-color"}
}

func (cmd plan) Env() []string {
	return []string{}
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

func (cmd importCmd) Env() []string {
	return []string{}
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

func (cmd renameProvider) Env() []string {
	return []string{}
}
