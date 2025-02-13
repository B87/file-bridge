package filesystem

import (
	"os"
	"path/filepath"
	"time"

	"github.com/B87/file-bridge/pkg/uri"
)

// Node represents a single file or directory in a filesystem.
type Node struct {
	URI   uri.URI
	IsDir bool

	Name    string
	Path    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
}

// NewNode creates a new Node from a URI.
func NewNode(uri uri.URI, isDir bool) Node {
	return Node{
		URI:   uri,
		IsDir: isDir,
		Name:  filepath.Base(uri.Path),
		Path:  uri.Path,
	}
}

// NodeTree is a hierarchical representation of one Node + any child nodes.
type NodeTree struct {
	Node     Node
	Children []NodeTree
}

// BuildTree constructs a hierarchical tree from a flat list of Nodes.
// It creates a dummy root ("") to hold the rest of the structure.
func BuildTree(nodes []Node) NodeTree {
	tree := NodeTree{
		Node: Node{
			Name:  "",
			Path:  "",
			IsDir: true,
		},
	}

	for _, nd := range nodes {
		tree.Children = append(tree.Children, NodeTree{
			Node: nd,
		})
	}

	return tree
}
