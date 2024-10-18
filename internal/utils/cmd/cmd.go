package cmd

import (
	"os"
	"os/exec"
)

func Verbose(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}
