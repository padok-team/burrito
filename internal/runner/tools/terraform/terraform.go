package terraform

import (
	"github.com/padok-team/burrito/internal/runner/tools/base"
)

type Terraform struct {
	base.BaseTool
}

func NewTerraform(execPath string) *Terraform {
	return &Terraform{
		BaseTool: base.BaseTool{
			ExecPath: execPath,
			ToolName: "terraform",
		},
	}
}
