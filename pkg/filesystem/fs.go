package filesystem

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"time"

	"github.com/B87/file-bridge/pkg/uri"
)

var (
	ErrFileCreate = errors.New("failed to create file")

	ErrFileList   = errors.New("failed to list files")
	ErrFileDelete = errors.New("failed to delete file")
	ErrFileCopy   = errors.New("failed to copy file")

	ErrNotFound      = fs.ErrNotExist
	ErrAlreadyExists = fs.ErrExist

	ErrDirNotEmpty = errors.New("directory not empty")
)

// FS is an abstraction for file system operations.
type FS interface {
	Reader
	Writer
}

type Reader interface {
	// Reader opens a file returning a reader that should be closed after use.
	Reader(ctx context.Context, fileName uri.URI) (io.ReadCloser, error)

	// Get gets a single file or directory.
	Get(ctx context.Context, path uri.URI) (Node, error)

	// List lists files in a path.
	List(ctx context.Context, path uri.URI, recursive bool) (NodeTree, error)
}

type Writer interface {
	// Writer creates a file returning a writer that should be closed after use.
	Writer(ctx context.Context, fileName uri.URI) (io.WriteCloser, error)

	// MkDir creates a directory
	MkDir(ctx context.Context, path uri.URI) (Node, error)

	// Delete deletes a file or directory.
	Delete(ctx context.Context, path uri.URI) error

	// Copy copies a file or directory on the same file system.
	Copy(ctx context.Context, src, dst uri.URI) error
}

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
		Name:  path.Base(uri.Path),
		Path:  uri.Path,
	}
}

// NodeTree is a hierarchical representation of one Node + any child nodes.
type NodeTree struct {
	Node     Node
	Children []NodeTree
}
