/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package datastore

import (
	"github.com/padok-team/burrito/internal/burrito"
	cmdUtils "github.com/padok-team/burrito/internal/utils/cmd"
	"github.com/spf13/cobra"
)

func BuildDatastoreCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datastore",
		Short: "cmd to use burrito's datastore",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If we reach this point, it means no subcommand was matched
			cmdUtils.UnsupportedCommand(cmd, args)
			return cmd.Help()
		},
	}
	cmd.AddCommand(buildDatastoreStartCmd(app))
	return cmd
}
