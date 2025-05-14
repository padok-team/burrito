package server

import (
	"github.com/padok-team/burrito/internal/burrito"
	cmdUtils "github.com/padok-team/burrito/internal/utils/cmd"
	"github.com/spf13/cobra"
)

func BuildServerCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "cmd to use burrito's server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If we reach this point, it means no subcommand was matched
			cmdUtils.UnsupportedCommand(cmd, args)
			return cmd.Help()
		},
	}
	cmd.AddCommand(buildServerStartCmd(app))
	return cmd
}
