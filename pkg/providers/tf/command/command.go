package command

type TerraformCommand interface {
	Command() []string
}
