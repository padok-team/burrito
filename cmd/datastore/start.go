/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package datastore

import (
	"github.com/padok-team/burrito/internal/burrito"
	"github.com/spf13/cobra"
)

func buildDatastoreStartCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Burrito Datastore",
		// Do not display usage on program error
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.StartRunner()
		},
	}
	cmd.Flags().StringVar(&app.Config.Server.Addr, "addr", ":8080", "addr the datastore listens on")
	return cmd
}
