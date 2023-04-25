package commits

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commits",
		Short: "Gather commit information and save to database",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Everything is good")
		},
	}
	return cmd
}
