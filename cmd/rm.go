package cmd

import (
	"github.com/spf13/cobra"

	"github.com/B87/file-bridge/pkg/filesys"
)

var rmCmd = &cobra.Command{
	Use:   "rm [file]",
	Short: "Remove a file",
	Long:  `Remove a file`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		recursive, _ := cmd.Flags().GetBool("recursive")
		verbose, _ := cmd.Flags().GetBool("verbose")
		logger := NewLogger(verbose)
		logger.Debug("Removing", args[0], "...")
		uri, err := filesys.ParseURI(args[0])
		fatalIfError(err)
		err = filesys.Delete(uri, recursive)
		fatalIfError(err)
	},
}

func init() {
	rmCmd.Flags().BoolP("recursive", "r", false, "Remove directories and their contents recursively")
	RootCmd.AddCommand(rmCmd)
}
