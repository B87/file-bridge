package filesys

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// LocalFS is a FileSystem implementation that uses the local disk.
type LocalFS struct{}

func NewLocalFS() *LocalFS { return &LocalFS{} }

func (LocalFS) Writer(name URI) (io.WriteCloser, error) {
	return os.Create(name.Path)
}

func (LocalFS) Reader(name URI) (io.ReadCloser, error) {
	return os.Open(name.Path)
}

/*
Use os package to delete file, mapping errors to custom errors

returns
  - ErrNotFound if file does not exist
  - ErrDirNotEmpty if directory is not empty and recursive is false
*/
func (l *LocalFS) Delete(name URI, recursive bool) error {
	info, err := os.Stat(name.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w : %s", ErrNotFound, name)
	}
	if info.IsDir() && !recursive {
		empty, err := l.IsEmpty(name)
		if err != nil {
			return err
		} else if !empty {
			return fmt.Errorf("%w : %s", ErrDirNotEmpty, name)
		}
	}
	if recursive {
		return os.RemoveAll(name.Path)
	} else {
		return os.Remove(name.Path)
	}
}

/*
Use os package to move file, mapping errors to custom errors

retruns:
  - ErrNotFound if source file does not exist
  - ErrDirNotEmpty if directory is not empty
  - ErrAlreadyExists if destination already exists
*/
func (l *LocalFS) Move(src, dst URI, recursive bool) error {
	srcInfo, err := os.Stat(src.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w : %s", ErrNotFound, src)
	}
	if srcInfo.IsDir() && !recursive {
		empty, err := l.IsEmpty(src)
		if err != nil {
			return err
		} else if !empty {
			return fmt.Errorf("%w : %s", ErrDirNotEmpty, src)
		}
	}
	dstInfo, err := os.Stat(dst.Path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if dstInfo.IsDir() {
		dst.Path = filepath.Join(dst.Path, filepath.Base(src.Path))
	}
	if exists, err := l.Exists(dst); exists {
		if err != nil {
			return err
		}
		return ErrAlreadyExists
	} else if !exists && err == nil {
		return os.Rename(src.Path, dst.Path)
	} else {
		return err
	}
}

/*
Use os package to copy file, mapping errors to custom errors

returns:
  - ErrNotFound if source file does not exist
*/
func (l *LocalFS) Copy(src, dst URI, recursive bool) error {
	srcInfo, err := os.Stat(src.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("src %w : %s", ErrNotFound, src)
	} else if err != nil {
		return err
	}
	dstInfo, err := os.Stat(dst.Path)
	if !errors.Is(err, fs.ErrNotExist) && !dstInfo.IsDir() {
		return fmt.Errorf("dst %w : %s", ErrAlreadyExists, dst)
	}

	if !srcInfo.IsDir() {
		return CopyFile(src, dst)
	} else {
		entries, err := os.ReadDir(src.Path)
		if err != nil {
			return err
		}
		for _, entry := range entries {

			srcPath := AppendURIPath(src, entry.Name())
			dstPath := AppendURIPath(dst, entry.Name())

			if entry.IsDir() {
				if err := l.Copy(srcPath, dstPath, true); err != nil {
					return err
				}
			} else {
				if err := CopyFile(srcPath, dstPath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func AppendURIPath(uri URI, path string) URI {
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
func (l *LocalFS) List(dir URI, recursive bool) ([]Node, error) {
	var files []Node
	if exists, err := l.Exists(dir); !exists && err == nil {
		return files, fmt.Errorf("%w : %s", ErrNotFound, dir)
	}
	err := filepath.Walk(dir.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory itself
		if path == dir.Path {
			return nil
		}
		if info.IsDir() {
			files = append(files, NewNode(path, true))
		} else {
			files = append(files, NewNode(path, false))
		}

		// Skip directories if not recursive mode
		if !recursive && info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return files, fmt.Errorf("%w : %v", ErrWalk, err)
	}
	return files, nil
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst URI) error {
	in, err := os.Open(src.Path)
	if err != nil {
		return err
	}
	defer in.Close()

	dstInfo, err := os.Stat(dst.Path)
	if err != nil {
		return err
	}
	if dstInfo.IsDir() {
		dst.Path = filepath.Join(dst.Path, filepath.Base(src.Path))
	}
	out, err := os.Create(dst.Path)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

var ErrWalk = errors.New("error walking the path")

func (l *LocalFS) Exists(path URI) (bool, error) {
	_, err := os.Stat(path.Path)
	if err == nil {
		return true, nil
	}
	// Check if file does not exist
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	// Return error if other error
	return false, err
}

/*
Use os package to get path as a Node.

Returns:
  - ErrNotFound if path does not exist
*/
func (l *LocalFS) Get(path URI) (Node, error) {
	info, err := os.Stat(path.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Node{FullPath: path.Path}, fmt.Errorf("%w : %s", ErrNotFound, path)
		}
		return Node{FullPath: path.Path}, err
	}
	return NewNode(path.Path, info.IsDir()), nil
}

func (l *LocalFS) IsEmpty(path URI) (bool, error) {
	if _, err := l.Exists(path); err != nil {
		return false, err
	}
	dir, err := os.Open(path.Path)
	if err != nil {
		return false, err
	}
	defer dir.Close()
	_, err = dir.Readdir(1)

	// If EOF is reached, the directory is empty
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

/*
Use os package to create a directory, recursive by default

Returns newly created Node or error:
  - ErrAlreadyExists if path already exists
*/
func (l *LocalFS) MkDir(path URI) (Node, error) {
	_, err := os.Stat(path.Path)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(path.Path, 0755); err != nil {
			return Node{FullPath: path.Path}, err
		}
		return NewNode(path.Path, true), nil
	} else if err != nil {
		return Node{FullPath: path.Path}, err
	} else {
		return Node{FullPath: path.Path}, ErrAlreadyExists
	}
}
