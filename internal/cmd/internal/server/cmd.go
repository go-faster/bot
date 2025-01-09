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
			app.Run(func(ctx context.Context, lg *zap.Logger, t *app.Telemetry) error {
				return runBot(ctx, t, lg.Named("bot"))
			})
		},
	}
	return cmd
}
