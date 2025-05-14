/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package runner

import (
	"github.com/padok-team/burrito/internal/burrito"
	cmdUtils "github.com/padok-team/burrito/internal/utils/cmd"
	"github.com/spf13/cobra"
)

func BuildRunnerCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runner",
		Short: "cmd to use burrito's runner",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If we reach this point, it means no subcommand was matched
			cmdUtils.UnsupportedCommand(cmd, args)
			return cmd.Help()
		},
	}
	cmd.AddCommand(buildRunnerStartCmd(app))
	return cmd
}
