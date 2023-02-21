package server

import (
	"github.com/padok-team/burrito/internal/burrito"
	"github.com/spf13/cobra"
)

func buildServerStartCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start burrito's server",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.StartServer()
			return nil
		},
	}

	cmd.Flags().StringVar(&app.Config.Server.Addr, "addr", ":8080", "addr the server listens on")

	return cmd
}
