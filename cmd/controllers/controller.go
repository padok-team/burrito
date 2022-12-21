/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package controllers

import (
	"github.com/padok-team/burrito/burrito"

	"github.com/spf13/cobra"
)

func BuildControllersCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.AddCommand(buildControllersStartCmd(app))
	return cmd
}
