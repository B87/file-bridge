package cmd

import (
	"github.com/spf13/cobra"

	"github.com/B87/file-bridge/pkg/uri"
)

var mkdirCmd = &cobra.Command{
	Use:   "mkdir [path]",
	Short: "Create a directory",
	Long:  `Create a directory`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			cmd.Println("Creating directory", args[0], "...")
		}
		uri, err := uri.Parse(args[0])
		fatalIfError(err)
		_, err = cmdManager.MkDir(cmd.Context(), uri)
		fatalIfError(err)
	},
}

func init() {
	mkdirCmd.Flags().BoolP("recursive", "r", false, "Create parent directories if they do not exist")
	RootCmd.AddCommand(mkdirCmd)
}
