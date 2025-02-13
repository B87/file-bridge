package cmd

import (
	"context"
	"log"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/B87/file-bridge/pkg/uri"
)

var mvCmd = &cobra.Command{
	Use:   "mv [source] [dest]",
	Short: "Move files from source to destination",
	Long:  `Move files from source to destination`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		recursive, _ := cmd.Flags().GetBool("recursive")
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			cmd.Println("Moving", args[0], "to", args[1], "...")
		}
		source, dest := validateArgs(args)
		srcURI, err := uri.Parse(source)
		fatalIfError(err)
		destURI, err := uri.Parse(dest)
		fatalIfError(err)

		bar := progressbar.DefaultBytes(
			-1,
			"Moving",
		)

		ctx := context.WithValue(cmd.Context(), "progressbar", bar)
		err = cmdManager.Move(ctx, srcURI, destURI, recursive)
		fatalIfError(err)

		if verbose {
			cmd.Println("\nFile moved")
		}
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
	mvCmd.Flags().BoolP("recursive", "r", false, "Move parent directories if they do not exist")
	RootCmd.AddCommand(mvCmd)
}
