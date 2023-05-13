package server

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/go-faster/sdk/app"
)

func Root() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run a go-faster bot server",
		Run: func(cmd *cobra.Command, args []string) {
			app.Run(func(ctx context.Context, lg *zap.Logger, m *app.Metrics) error {
				return runBot(ctx, m, lg.Named("bot"))
			})
		},
	}
	return cmd
}
