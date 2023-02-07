package server

import (
	"github.com/padok-team/burrito/internal/burrito"
	"github.com/spf13/cobra"
)

func BuildServerCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "cmd to use burrito's server",
	}
	cmd.AddCommand(buildServerStartCmd(app))
	return cmd
}
