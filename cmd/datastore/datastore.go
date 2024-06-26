/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package datastore

import (
	"github.com/padok-team/burrito/internal/burrito"
	"github.com/spf13/cobra"
)

func BuildDatastoreCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datastore",
		Short: "cmd to use burrito's datastore",
	}
	cmd.AddCommand(buildDatastoreStartCmd(app))
	return cmd
}
