package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/B87/file-bridge/pkg/filesys"
)

var mvCmd = &cobra.Command{
	Use:   "mv [source] [dest]",
	Short: "Move files from source to destination",
	Long:  `Move files from source to destination`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		recursive, _ := cmd.Flags().GetBool("recursive")
		verbose, _ := cmd.Flags().GetBool("verbose")
		logger := NewLogger(verbose)

		source, dest := validateArgs(args)
		srcURI, err := filesys.ParseURI(source)
		fatalIfError(err)
		destURI, err := filesys.ParseURI(dest)
		fatalIfError(err)

		filesys.Move(srcURI, destURI, recursive)
		logger.Debug("File moved")
	},
}

func validateArgs(args []string) (string, string) {
	if len(args) != 2 {
		log.Fatal("source and destination are required")
	}
	source := args[0]
	if source == "" {
		log.Fatal("source is required")
	}

	destination := args[1]
	if destination == "" {
		log.Fatal("destination is required")
	}
	return source, destination
}

func init() {
	RootCmd.AddCommand(mvCmd)
}
