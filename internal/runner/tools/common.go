package tools

type BaseExec interface {
	Init(string) error
	Plan(string) error
	Apply(string) error
	Show(string, string) ([]byte, error)
	TenvName() string
	GetExecPath() string
}
