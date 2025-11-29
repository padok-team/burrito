package tools

type BaseExec interface {
	Init(string) error
	Plan(string) error
	PlanDestroy(string) error
	Apply(string) error
	Show(string, string) ([]byte, error)
	TenvName() string
	GetExecPath() string
}
