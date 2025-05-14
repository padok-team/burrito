package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

func Verbose(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}

func UnsupportedCommand(subcommand string, args []string) {
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "Error: unknown %s subcommand: %s\n", subcommand, args[0])
	}
	fmt.Fprintf(os.Stderr, "Run 'burrito %s --help' for usage", subcommand)
	os.Exit(2)
}
