/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package controllers

import (
	"github.com/padok-team/burrito/internal/burrito"
	"github.com/spf13/cobra"
)

func buildControllersStartCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Burrito controllers",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.StartController()
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&app.Config.Controller.Types, "types", []string{"layer", "repository"}, "list of controllers to start")

	cmd.Flags().StringVar(&app.Config.Controller.Timers.DriftDetection, "drift-detection-period", "20m", "period between two plans. Must end with s, m or h.")

	return cmd
}
