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

	cmd.PersistentFlags().StringVar(&app.Config.Redis.Hostname, "redis-host", "burrito-redis.burrito-system", "the redis host to connect to")
	cmd.PersistentFlags().IntVar(&app.Config.Redis.Port, "redis-port", 6379, "the port of the redis to connect to")
	cmd.PersistentFlags().StringVar(&app.Config.Redis.Password, "redis-password", "", "the redis password")
	cmd.PersistentFlags().IntVar(&app.Config.Redis.Database, "redis-database", 0, "the redis database")

	cmd.AddCommand(controllers.BuildControllersCmd(app))
	cmd.AddCommand(runner.BuildRunnerCmd(app))
	cmd.AddCommand(server.BuildServerCmd(app))
	cmd.AddCommand(buildVersionCmd())
	return cmd
}
