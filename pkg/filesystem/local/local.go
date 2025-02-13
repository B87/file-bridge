package local

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/B87/file-bridge/pkg/filesystem"
	"github.com/B87/file-bridge/pkg/uri"
)

var ErrWalk = errors.New("error walking the path")

// LocalFS is a FileSystem implementation that uses the local disk.
type LocalFS struct{}

func NewLocalFS() *LocalFS { return &LocalFS{} }

func (LocalFS) Writer(ctx context.Context, name uri.URI) (io.WriteCloser, error) {
	return os.Create(name.Path)
}

func (LocalFS) Reader(ctx context.Context, name uri.URI) (io.ReadCloser, error) {
	return os.Open(name.Path)
}

/*
Use os package to delete file, mapping errors to custom errors

returns
  - ErrNotFound if file does not exist
  - ErrDirNotEmpty if directory is not empty and recursive is false
*/
func (l *LocalFS) Delete(ctx context.Context, name uri.URI) error {
	err := os.RemoveAll(name.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w : %s", filesystem.ErrFileDelete, name)
	}
	return err
}

/*
Use os package to copy file, mapping errors to custom errors

returns:
  - ErrNotFound if source file does not exist
*/
func (l *LocalFS) Copy(ctx context.Context, src, dst uri.URI) error {
	srcInfo, err := os.Stat(src.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("src %w : %s", filesystem.ErrNotFound, src)
	} else if err != nil {
		return err
	}
	dstInfo, err := os.Stat(dst.Path)
	if !errors.Is(err, fs.ErrNotExist) && !dstInfo.IsDir() {
		return fmt.Errorf("dst %w : %s", filesystem.ErrAlreadyExists, dst)
	}

	// Create destination directory if it doesn't exist
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(dst.Path), 0755); err != nil {
			return err
		}
	}

	if !srcInfo.IsDir() {
		return l.copyLocalFile(src, dst)
	} else {
		// Create the destination directory for recursive copy
		if err := os.MkdirAll(dst.Path, 0755); err != nil {
			return err
		}

		entries, err := os.ReadDir(src.Path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			srcPath := appendURIPath(src, entry.Name())
			dstPath := appendURIPath(dst, entry.Name())
			if entry.IsDir() {
				if err := l.Copy(ctx, srcPath, dstPath); err != nil {
					return err
				}
			} else {
				if err := l.copyLocalFile(srcPath, dstPath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func appendURIPath(uri uri.URI, path string) uri.URI {
	uri.Path = filepath.Join(uri.Path, path)
	return uri
}

/*
Use filepath.Walk to list files, mapping errors to custom errors

# If dir is a file, returns a list containing only that file as a node

returns slice of nodes in dir and an error:
  - ErrNotFound if dir does not exist
  - ErrWalk if error walking the path
*/
func (l *LocalFS) List(ctx context.Context, dir uri.URI, recursive bool) (filesystem.NodeTree, error) {
	var tree filesystem.NodeTree
	// Check if path exists
	info, err := os.Stat(dir.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return tree, fmt.Errorf("%w : %s", filesystem.ErrNotFound, dir)
		}
		return tree, err
	}

	// If it's a file, return just that file
	if !info.IsDir() {
		node := filesystem.NewNode(dir, false)
		return filesystem.NodeTree{
			Node:     node,
			Children: []filesystem.NodeTree{},
		}, nil
	}

	// Handle directory listing
	entries, err := os.ReadDir(dir.Path)
	if err != nil {
		return tree, err
	}

	children := make([]filesystem.NodeTree, 0, len(entries))
	for _, entry := range entries {
		childPath := filepath.Join(dir.Path, entry.Name())
		childURI := uri.URI{
			Scheme: dir.Scheme,
			Path:   childPath,
		}

		childNode := filesystem.NewNode(childURI, entry.IsDir())
		childTree := filesystem.NodeTree{Node: childNode}

		if recursive && entry.IsDir() {
			subTree, err := l.List(ctx, childURI, recursive)
			if err != nil {
				return tree, err
			}
			childTree.Children = subTree.Children
		}

		children = append(children, childTree)
	}

	return filesystem.NodeTree{
		Node:     filesystem.NewNode(dir, true),
		Children: children,
	}, nil
}

// copyLocalFile copies a single file from src to dst
func (l *LocalFS) copyLocalFile(src, dst uri.URI) error {
	in, err := os.Open(src.Path)
	if err != nil {
		return err
	}
	defer in.Close()

	// Check if destination is a directory only if it exists
	if info, err := os.Stat(dst.Path); err == nil && info.IsDir() {
		dst.Path = filepath.Join(dst.Path, filepath.Base(src.Path))
	}

	out, err := os.Create(dst.Path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err // No need to call Close() again as defer will handle it
}

/*
Use os package to get path as a Node.

Returns:
  - ErrNotFound if path does not exist
*/
func (l *LocalFS) Get(ctx context.Context, path uri.URI) (filesystem.Node, error) {
	info, err := os.Stat(path.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return filesystem.Node{URI: path}, fmt.Errorf("%w : %s", filesystem.ErrNotFound, path)
		}
		return filesystem.Node{URI: path}, err
	}
	return filesystem.NewNode(path, info.IsDir()), nil
}

/*
Use os package to create a directory, recursive by default

Returns newly created Node or error:
  - ErrAlreadyExists if path already exists
*/
func (l *LocalFS) MkDir(ctx context.Context, path uri.URI) (filesystem.Node, error) {
	node := filesystem.NewNode(path, true)

	// First check if the exact path already exists
	if _, err := os.Stat(path.Path); err == nil {
		// Path exists - return error
		return node, filesystem.ErrAlreadyExists
	} else if !errors.Is(err, fs.ErrNotExist) {
		// Some other error occurred
		return node, err
	}

	// Path doesn't exist, create it and all necessary parent directories
	if err := os.MkdirAll(path.Path, 0755); err != nil {
		return node, err
	}

	return node, nil
}
