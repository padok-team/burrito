package cmd

import (
	"fmt"

	"github.com/padok-team/burrito/internal/version"
	"github.com/spf13/cobra"
)

func buildVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.BuildVersion())
		},
	}
	return cmd
}
