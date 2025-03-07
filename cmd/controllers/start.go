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

	defaultDriftDetectionTimer, _ := time.ParseDuration("4h")
	defaultOnErrorTimer, _ := time.ParseDuration("10s")
	defaultWaitActionTimer, _ := time.ParseDuration("5s")
	defaultFailureGracePeriod, _ := time.ParseDuration("15s")
	defaultRepositorySyncTimer, _ := time.ParseDuration("5m")

	cmd.Flags().StringSliceVar(&app.Config.Controller.Namespaces, "namespaces", []string{"burrito-system"}, "list of namespaces to watch")
	cmd.Flags().StringArrayVar(&app.Config.Controller.Types, "types", []string{"layer", "run", "pullrequest"}, "list of controllers to start")
	cmd.Flags().DurationVar(&app.Config.Controller.Timers.DriftDetection, "drift-detection-period", defaultDriftDetectionTimer, "period between two plans. Must end with s, m or h.")
	cmd.Flags().DurationVar(&app.Config.Controller.Timers.RepositorySync, "repository-sync-period", defaultRepositorySyncTimer, "period between two repository sync. Must end with s, m or h.")
	cmd.Flags().DurationVar(&app.Config.Controller.Timers.OnError, "on-error-period", defaultOnErrorTimer, "period between two runners launch when an error occurred in the controllers. Must end with s, m or h.")
	cmd.Flags().DurationVar(&app.Config.Controller.Timers.WaitAction, "wait-action-period", defaultWaitActionTimer, "period between two runners when a layer is locked. Must end with s, m or h.")
	cmd.Flags().DurationVar(&app.Config.Controller.Timers.FailureGracePeriod, "failure-grace-period", defaultFailureGracePeriod, "initial time before retry, goes exponential function of number failure. Must end with s, m or h.")
	cmd.Flags().IntVar(&app.Config.Controller.TerraformMaxRetries, "terraform-max-retries", 5, "default number of retries for terraform actions (can be overriden in CRDs)")
	cmd.Flags().IntVar(&app.Config.Controller.MaxConcurrentReconciles, "max-concurrent-reconciles", 1, "maximum number of concurrent reconciles")
	cmd.Flags().IntVar(&app.Config.Controller.MaxConcurrentRunnerPods, "max-runner-pods", 0, "maximum number of concurrent runner pods")
	cmd.Flags().BoolVar(&app.Config.Controller.LeaderElection.Enabled, "leader-election", true, "whether leader election is enabled or not, default to true")
	cmd.Flags().StringVar(&app.Config.Controller.LeaderElection.ID, "leader-election-id", "6d185457.terraform.padok.cloud", "lease id used for leader election")
	cmd.Flags().StringVar(&app.Config.Controller.HealthProbeBindAddress, "health-probe-bind-address", ":8081", "address to bind the metrics server embedded in the controllers")
	cmd.Flags().StringVar(&app.Config.Controller.MetricsBindAddress, "metrics-bind-address", ":8080", "address to bind the health probe server embedded in the controllers")
	cmd.Flags().IntVar(&app.Config.Controller.KubernetesWebhookPort, "kubernetes-webhook-port", 9443, "port used by the validating webhook server embedded in the controllers")
	return cmd
}
