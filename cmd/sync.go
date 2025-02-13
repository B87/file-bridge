package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(syncCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync [uri] [uri]",
	Short: "Sync two URIs",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Context()
		_, _ = cmd.Flags().GetBool("verbose")

	},
}
