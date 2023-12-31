package cmd

import (
	"github.com/spf13/cobra"

	"github.com/B87/file-bridge/pkg/filesys"
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
		logger := NewLogger(verbose)
		source, dest := validateArgs(args)
		srcURI, err := filesys.ParseURI(source)
		fatalIfError(err)
		logger.Debugf("Source URI: %s", srcURI)

		dstURI, err := filesys.ParseURI(dest)
		fatalIfError(err)
		logger.Debugf("Destination URI: %s", dstURI)

		err = filesys.Copy(srcURI, dstURI, recursive)
		fatalIfError(err)
	},
}

func init() {
	cpCMD.Flags().BoolP("recursive", "r", false, "Copy directories recursively")
	RootCmd.AddCommand(cpCMD)
}
