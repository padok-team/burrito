package webhook

import (
	"github.com/padok-team/burrito/internal/burrito"
	"github.com/spf13/cobra"
)

func BuildWebhookCmd(app *burrito.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "cmd to use burrito's webhook",
	}
	cmd.AddCommand(buildWebhookStartCmd(app))
	return cmd
}
