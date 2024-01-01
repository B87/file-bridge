package filesys

import (
	"errors"
	"fmt"
	"io"
	"path"
)

/*
Copy copies a file from one filesystem to another.

If both filesystems are the same, use the filesystem's copy method.

If the filesystems are different, perform a manual copy:
  - get the source file
  - create the destination file
*/
func Copy(src, dst URI, recursive bool) error {
	srcFS := SchemeFS(src.Scheme)
	dstFS := SchemeFS(dst.Scheme)
	err := connectFilesystems(srcFS, dstFS)
	if err != nil {
		return err
	}
	defer disconnectFilesystems(srcFS, dstFS)

	// If the destination is a directory, add the source file name to the destination path
	srcNode, err := srcFS.Get(src)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("failed to get source file %s: %w", src.String(), err)
	} else if errors.Is(err, ErrNotFound) {
		return fmt.Errorf("source file %s does not exist", src.String())
	}
	if srcNode.IsDir {
		dst.Path = path.Join(dst.Path, srcNode.URI.Name)
		dstFS.MkDir(dst)
	}
	// If the filesystems are the same, use the filesystem's copy method
	// we might get a better performance using the nateive copy method if exists
	if srcFS == dstFS {
		return srcFS.Copy(src, dst, recursive)
	} else {
		nodes, err := List(src, recursive)
		if err != nil {
			return err
		}
		for _, node := range nodes {
			if !node.IsDir {
				err := CopyFile(node.URI, dst, srcFS, dstFS)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func CopyFile(src, dst URI, srcFS, dstFS FS) error {
	srcFile, err := srcFS.Reader(src)
	if err != nil {
		return err
	}
	// Modify the destination path to include the file name
	dst.Path = path.Join(dst.Path, src.Name)
	dstFile, err := dstFS.Writer(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	err = dstFile.Close()
	if err != nil {
		return err
	}
	err = srcFile.Close()
	if err != nil {
		return err
	}
	return nil
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
