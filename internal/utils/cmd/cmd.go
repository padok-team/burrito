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
	if err := cmd.Help(); err != nil {
		fmt.Fprintf(os.Stderr, "Error displaying help: %v\n", err)
		os.Exit(1)
	}
	os.Exit(2)
}
