// Package command is an interface for the Terraform command
package command

type TerraformCommand interface {
	Command() []string
	Env() []string
}
