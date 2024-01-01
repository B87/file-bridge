package cmd

import (
	"github.com/spf13/cobra"

	"github.com/B87/file-bridge/pkg/filesys"
)

var ListCmd = &cobra.Command{
	Use:   "ls [directory]",
	Short: "List files in a directory",
	Long:  `List files in a directory`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var dir string
		if len(args) == 0 {
			dir = "."
		} else {
			dir = args[0]
		}
		recursive, _ := cmd.Flags().GetBool("recursive")
		verbose, _ := cmd.Flags().GetBool("verbose")

		logger := NewLogger(verbose)
		logger.PrintDebugGlobal()

		uri, err := filesys.ParseURI(dir)
		fatalIfError(err)
		files, err := filesys.List(uri, recursive)
		fatalIfError(err)

		logger.Debug("Listing files in", uri.Path, "from file system", uri.Scheme, "...\n")
		for _, file := range files {
			logger.Print(file.URI.Path)
		}

	},
}

func init() {
	ListCmd.Flags().BoolP("recursive", "r", false, "List files recursively")
	RootCmd.AddCommand(ListCmd)
}
