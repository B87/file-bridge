package filesys

import "errors"

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
	if srcFS == dstFS {
		return srcFS.Copy(src, dst, recursive)
	} else {
		// TODO: implement manual copy
		return errors.New("cannot copy between different filesystems yet")
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
	return srcFS.Delete(src, recursive)
}

/*
Lists the contents of a directory. If recursive is true, lists recursively.
*/
func List(path URI, recursive bool) ([]Node, error) {
	fs := SchemeFS(path.Scheme)
	return fs.List(path, recursive)
}

// Delete deletes a file or directory
func Delete(path URI, recursive bool) error {
	fs := SchemeFS(path.Scheme)
	return fs.Delete(path, recursive)
}

// MkDir creates an empty directory
func MkDir(path URI) (Node, error) {
	fs := SchemeFS(path.Scheme)
	return fs.MkDir(path)
}
