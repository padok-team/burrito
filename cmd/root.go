/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/padok-team/burrito/cmd/controllers"
	"github.com/padok-team/burrito/cmd/runner"
	"github.com/padok-team/burrito/cmd/server"
	"github.com/padok-team/burrito/internal/burrito"

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

	cmd.Flags().StringVar(&app.Config.Redis.URL, "redis-url", "redis:6379", "the redis URL to connect to")
	cmd.Flags().StringVar(&app.Config.Redis.Password, "redis-password", "", "the redis password")
	cmd.Flags().IntVar(&app.Config.Redis.Database, "redis-password", 0, "the redis database")

	cmd.AddCommand(controllers.BuildControllersCmd(app))
	cmd.AddCommand(runner.BuildRunnerCmd(app))
	cmd.AddCommand(server.BuildServerCmd(app))
	cmd.AddCommand(buildVersionCmd())
	return cmd
}
