package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/B87/file-bridge/pkg/filesystem"
	"github.com/B87/file-bridge/pkg/uri"
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

		uriVal, err := uri.Parse(dir)
		fatalIfError(err)
		files, err := cmdManager.List(cmd.Context(), uriVal, recursive)
		fatalIfError(err)

		printList(cmd, files)
	},
}

func init() {
	ListCmd.Flags().BoolP("recursive", "r", false, "List files recursively")
	RootCmd.AddCommand(ListCmd)
}

func printList(cmd *cobra.Command, tree filesystem.NodeTree) {
	// Skip printing the root node (current directory)
	for _, child := range tree.Children {
		printNode(child, "", true)
	}
}

func printNode(tree filesystem.NodeTree, prefix string, isLast bool) {
	// Create branch symbols
	branch := "├── "
	if isLast {
		branch = "└── "
	}

	// Print the current node
	nodePrefix := prefix + branch
	if tree.Node.IsDir {
		fmt.Printf("%s%s/\n", nodePrefix, tree.Node.Name)
	} else {
		fmt.Printf("%s%s\n", nodePrefix, tree.Node.Name)
	}

	// Print children with updated prefix
	childPrefix := prefix
	if isLast {
		childPrefix += "    " // 4 spaces when last
	} else {
		childPrefix += "│   " // vertical line + 3 spaces
	}

	for i, child := range tree.Children {
		isLastChild := i == len(tree.Children)-1
		printNode(child, childPrefix, isLastChild)
	}
}
