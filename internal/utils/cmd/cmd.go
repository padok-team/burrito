package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func Verbose(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}

func UnsupportedCommand(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "Error: unknown %s subcommand: %s\n", cmd.Use, args[0])
	}
	fmt.Fprintf(os.Stderr, "Run 'burrito %s --help' for usage\n", cmd.Use)
	cmd.Help()
	os.Exit(2)
}
