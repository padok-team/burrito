/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package runner

import (
	"github.com/padok-team/burrito/internal/burrito"
	"github.com/spf13/cobra"
)

func buildRunnerStartCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Burrito runner",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.StartRunner()
			return nil
		},
	}
	return cmd
}
