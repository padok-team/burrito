/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package runner

import (
	"github.com/padok-team/burrito/internal/burrito"
	"github.com/spf13/cobra"
)

func buildRunnerStartCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Burrito runner",
		// Do not display usage on program error
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.StartRunner()
		},
	}

	cmd.Flags().StringVar(&app.Config.Runner.SSHKnownHostsConfigMapName, "ssh-known-hosts-cm-name", "burrito-ssh-known-hosts", "configmap name to get known hosts file from")
	cmd.Flags().StringVar(&app.Config.Runner.RunnerBinaryPath, "runner-binary-path", "/runner/bin", "binary path where the runner can expect to find terraform or terragrunt binaries")
	return cmd
}
