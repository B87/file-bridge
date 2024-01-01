package filesys

import (
	"errors"
	"io"
	"path"
	"strings"
)

/*
Copy copies a file from one filesystem to another.

If both filesystems are the same, use the filesystem's copy method.

If the filesystems are different, perform a manual copy:
  - get the source file
  - create the destination file
*/
func Copy(src, dst URI, recursive bool) error {

	// If the destination is a directory, append the source file name to the destination path
	if strings.HasSuffix(dst.Path, "/") {
		dst.Path = path.Join(dst.Path, src.Name)
	}

	srcFS := SchemeFS(src.Scheme)
	dstFS := SchemeFS(dst.Scheme)
	err := connectFilesystems(srcFS, dstFS)
	if err != nil {
		return err
	}
	defer disconnectFilesystems(srcFS, dstFS)
	if srcFS == dstFS {
		return srcFS.Copy(src, dst, recursive)
	} else {
		nodes, err := List(src, recursive)
		if err != nil {
			return err
		}
		for _, node := range nodes {
			if !node.IsDir {
				// Copy object from src to dst, if
				srcFile, err := srcFS.Reader(node.URI)
				if err != nil {
					return err
				}

				dstFile, err := dstFS.Writer(node.URI)
				if err != nil {
					return err
				}
				_, err = io.Copy(dstFile, srcFile)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

/*
Move moves a file from one filesystem to another.

Is implemented as a copy [src] [dst] followed by a delete [src].
*/
func Move(src, dst URI, recursive bool) error {
	err := Copy(src, dst, recursive)
	if err != nil {
		return err
	}
	srcFS := SchemeFS(src.Scheme)
	err = connectFilesystems(srcFS)
	if err != nil {
		return err
	}
	defer disconnectFilesystems(srcFS)
	return srcFS.Delete(src, recursive)
}

/*
Lists the contents of a directory. If recursive is true, lists recursively.
*/
func List(path URI, recursive bool) ([]Node, error) {
	fs := SchemeFS(path.Scheme)
	err := connectFilesystems(fs)
	if err != nil {
		return []Node{}, err
	}
	defer disconnectFilesystems(fs)
	return fs.List(path, recursive)
}

// Delete deletes a file or directory
func Delete(path URI, recursive bool) error {
	fs := SchemeFS(path.Scheme)
	err := connectFilesystems(fs)
	if err != nil {
		return err
	}
	defer disconnectFilesystems(fs)
	return fs.Delete(path, recursive)
}

// MkDir creates an empty directory
func MkDir(path URI) (Node, error) {
	fs := SchemeFS(path.Scheme)
	err := connectFilesystems(fs)
	if err != nil {
		return Node{}, err
	}
	defer disconnectFilesystems(fs)
	return fs.MkDir(path)
}

func connectFilesystems(filesystems ...FS) error {
	for _, fs := range filesystems {
		if err := fs.Connect(); err != nil {
			return errors.Join(ErrConnecting, err)
		}
	}
	return nil
}

func disconnectFilesystems(filesystems ...FS) error {
	for _, fs := range filesystems {
		if err := fs.Disconnect(); err != nil {
			return errors.Join(ErrDisconnecting, err)
		}
	}
	return nil
}
