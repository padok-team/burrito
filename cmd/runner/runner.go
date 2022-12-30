/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package runner

import (
	"github.com/padok-team/burrito/internal/burrito"
	"github.com/spf13/cobra"
)

func BuildRunnerCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runner",
		Short: "cmd to use burrito's runner",
	}
	cmd.AddCommand(buildRunnerStartCmd(app))
	return cmd
}
