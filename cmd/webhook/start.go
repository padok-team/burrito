package webhook

import (
	"github.com/padok-team/burrito/burrito"
	"github.com/spf13/cobra"
)

func buildWebhookStartCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Burrito webhook",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.StartWebhook()
			return nil
		},
	}
	return cmd
}
