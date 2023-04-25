package job

import (
	"github.com/spf13/cobra"

	"github.com/go-faster/bot/internal/cmd/internal/job/commits"
)

func Root() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Run a job",
	}
	cmd.AddCommand(
		commits.Root(),
	)
	return cmd
}
