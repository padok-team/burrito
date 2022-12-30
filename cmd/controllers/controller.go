/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package controllers

import (
	"github.com/padok-team/burrito/internal/burrito"

	"github.com/spf13/cobra"
)

func BuildControllersCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controllers",
		Short: "cmd to use burrito's controllers",
	}
	cmd.AddCommand(buildControllersStartCmd(app))
	return cmd
}
