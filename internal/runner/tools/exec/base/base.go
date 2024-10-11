package base

import e "github.com/padok-team/burrito/internal/runner/tools/exec"

type BaseExec interface {
	e.Exec
	GetExecPath() string
}
