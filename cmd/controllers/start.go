/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package controllers

import (
	"time"

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

	defaultDriftDetectionTimer, _ := time.ParseDuration("20m")
	defaultOnErrorTimer, _ := time.ParseDuration("1m")
	defaultWaitActionTimer, _ := time.ParseDuration("1m")

	cmd.Flags().StringSliceVar(&app.Config.Controller.Types, "types", []string{"layer", "repository"}, "list of controllers to start")

	cmd.Flags().DurationVar(&app.Config.Controller.Timers.DriftDetection, "drift-detection-period", defaultDriftDetectionTimer, "period between two plans. Must end with s, m or h.")
	cmd.Flags().DurationVar(&app.Config.Controller.Timers.OnError, "on-error-period", defaultOnErrorTimer, "period between two runners launch when an error occurred. Must end with s, m or h.")
	cmd.Flags().DurationVar(&app.Config.Controller.Timers.WaitAction, "wait-action-period", defaultWaitActionTimer, "period between two runners when a layer is locked. Must end with s, m or h.")
	cmd.Flags().BoolVar(&app.Config.Controller.LeaderElection.Enabled, "leader-election", true, "whether leader election is enabled or not, default to true")
	cmd.Flags().StringVar(&app.Config.Controller.LeaderElection.ID, "leader-election-id", "6d185457.terraform.padok.cloud", "lease id used for leader election")
	cmd.Flags().StringVar(&app.Config.Controller.HealthProbeBindAddress, "health-probe-bind-address", ":8081", "address to bind the metrics server embedded in the controllers")
	cmd.Flags().StringVar(&app.Config.Controller.MetricsBindAddress, "metrics-bind-address", ":8080", "address to bind the health probe server embedded in the controllers")
	cmd.Flags().IntVar(&app.Config.Controller.KubernetesWehbookPort, "kubernetes-webhook-port", 9443, "port used by the validating webhook server embedded in the controllers")
	return cmd
}
