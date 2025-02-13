package cmd

import (
	"github.com/spf13/cobra"

	"github.com/B87/file-bridge/pkg/uri"
)

var cpCMD = &cobra.Command{
	Use:   "cp [source] [destination]",
	Short: "Copy files from source to destination",
	Long: `
  Copy files from source to destination:

  filer cp tmp/file.txt tmp/file2.txt
  filer cp tmp/file.txt gs://bucket/file.txt

  filer cp -r tmp gs://bucket
  filer cp -r gs://bucket tmp
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		recursive, _ := cmd.Flags().GetBool("recursive")
		if verbose {
			cmd.Println("Copying", args[0], "to", args[1], "...")
		}
		source, dest := validateArgs(args)
		srcURI, err := uri.Parse(source)
		fatalIfError(err)
		if verbose {
			cmd.Printf("Source URI: %s\n", srcURI)
		}

		dstURI, err := uri.Parse(dest)
		fatalIfError(err)
		if verbose {
			cmd.Printf("Destination URI: %s\n", dstURI)
		}

		err = cmdManager.Copy(cmd.Context(), srcURI, dstURI, recursive)
		fatalIfError(err)
		if verbose {
			cmd.Println("File copied")
		}
	},
}

func init() {
	cpCMD.Flags().BoolP("recursive", "r", false, "Copy directories recursively")
	RootCmd.AddCommand(cpCMD)
}
