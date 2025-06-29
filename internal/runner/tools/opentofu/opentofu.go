package opentofu

import (
	"github.com/padok-team/burrito/internal/runner/tools/base"
)

type OpenTofu struct {
	base.BaseTool
}

func NewOpenTofu(execPath string) *OpenTofu {
	return &OpenTofu{
		BaseTool: base.BaseTool{
			ExecPath: execPath,
			ToolName: "tofu",
		},
	}
}
