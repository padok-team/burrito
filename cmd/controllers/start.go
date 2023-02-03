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
			err := app.StartController()
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&app.Config.Controller.Types, "types", []string{"layer", "repository"}, "list of controllers to start")

	cmd.Flags().StringVar(&app.Config.Controller.Timers.DriftDetection, "drift-detection-period", "20m", "period between two plans. Must end with s, m or h.")
	cmd.Flags().StringVar(&app.Config.Controller.Timers.OnError, "on-error-period", "1m", "period between two runners launch when an error occured. Must end with s, m or h.")
	cmd.Flags().StringVar(&app.Config.Controller.Timers.WaitAction, "wait-action-period", "1m", "period between two runners when a layer is locked. Must end with s, m or h.")

	return cmd
}
