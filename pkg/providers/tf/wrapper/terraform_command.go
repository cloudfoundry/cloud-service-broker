package wrapper

type TerraformCommand interface {
	Command() []string
}

type InitCommand struct{}

func (InitCommand) Command() []string {
	return []string{"init", "-no-color"}
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
