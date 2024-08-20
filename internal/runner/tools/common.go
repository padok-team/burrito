package tools

type TerraformExec interface {
	Init(string) error
	Plan(string) error
	Apply(string) error
	Show(string, string) ([]byte, error)
}
