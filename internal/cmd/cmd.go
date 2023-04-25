package cmd

import (
	"github.com/spf13/cobra"

	"github.com/go-faster/bot/internal/cmd/internal/job"
	"github.com/go-faster/bot/internal/cmd/internal/server"
)

func Root() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bot",
		Short: "The go-faster automation around GitHub and Telegram",
	}
	cmd.AddCommand(
		job.Root(),
		server.Root(),
	)
	return cmd
}
