package cmd

import (
	"github.com/spf13/cobra"

	"github.com/B87/file-bridge/pkg/uri"
)

var rmCmd = &cobra.Command{
	Use:   "rm [uri]",
	Short: "Remove a URI and all its contents",
	Long:  `Remove a URI and all its contents`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			cmd.Println("Removing", args[0], "...")
		}
		uri, err := uri.Parse(args[0])
		fatalIfError(err)
		err = cmdManager.Delete(cmd.Context(), uri)
		fatalIfError(err)
	},
}

func init() {
	rmCmd.Flags().BoolP("recursive", "r", false, "Remove directories and their contents recursively")
	RootCmd.AddCommand(rmCmd)
}
