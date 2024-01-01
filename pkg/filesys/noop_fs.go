package filesys

import "io"

// NoopFS is a FileSystem implementation that does nothing.
type NoopFS struct{}

func (NoopFS) Connect() error    { return nil }
func (NoopFS) Disconnect() error { return nil }

func (NoopFS) Writer(name URI) (io.WriteCloser, error) {
	if name.Path == "badFile.jpg" {
		return NoopFile{io.Discard}, nil
	}
	return nil, ErrFileCreate
}
func (NoopFS) Reader(name URI) (io.ReadCloser, error)          { return nil, ErrFileOpen }
func (NoopFS) Delete(name URI, recursive bool) error           { return nil }
func (NoopFS) Move(oldName, newName URI, recursive bool) error { return nil }
func (NoopFS) Copy(oldName, newName URI, recursive bool) error { return nil }
func (NoopFS) List(dir URI, recursive bool) ([]Node, error)    { return []Node{}, nil }
func (NoopFS) Exists(path URI) (bool, error)                   { return true, nil }
func (NoopFS) Get(path URI) (Node, error)                      { return Node{}, nil }
func (NoopFS) MkDir(path URI) (Node, error)                    { return Node{}, nil }

// NoopFile is a WriteCloser implementation that returns nothing.
type NoopFile struct{ io.Writer }

func (NoopFile) Close() error { return ErrFileClose }
