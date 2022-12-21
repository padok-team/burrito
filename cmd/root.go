/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/padok-team/burrito/burrito"
	"github.com/padok-team/burrito/cmd/controllers"
	"github.com/padok-team/burrito/cmd/runner"

	"github.com/spf13/cobra"
)

func New(app *burrito.App) *cobra.Command {
	return buildBurritoCmd(app)
}

func buildBurritoCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burrito",
		Short: "burrito is a TACoS",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return app.Config.Load(cmd.Flags())
		},
	}

	cmd.AddCommand(controllers.BuildControllersCmd(app))
	cmd.AddCommand(runner.BuildRunnerCmd(app))
	return cmd
}
