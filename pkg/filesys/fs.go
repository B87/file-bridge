package filesys

import (
	"errors"
	"io"
	"path"
	"regexp"
)

var (
	ErrFileCreate    = errors.New("failed to create file")
	ErrFileClose     = errors.New("failed to close file")
	ErrFileOpen      = errors.New("failed to open file")
	ErrorFilMove     = errors.New("failed to move file")
	ErrFileList      = errors.New("failed to list files")
	ErrNotFound      = errors.New("file not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrDirNotEmpty   = errors.New("directory not empty")

	ErrConnecting    = errors.New("failed to connect filesystem")
	ErrDisconnecting = errors.New("failed to disconnect filesystem")
)

// FS is an abstraction for file system operations.
type FS interface {
	/*
		Connect connects to the file system.
		It should be called before any other method since some file systems
		may require an initialization.
	*/
	Connect() error
	// Disconnect closes the connection from the file system.
	Disconnect() error

	// Create creates a file.
	Writer(fileName URI) (io.WriteCloser, error)
	// Open opens a file.
	Reader(fileName URI) (io.ReadCloser, error)

	// Delete deletes a file or directory.
	Delete(path URI, recursive bool) error
	// Copy copies a file or directory on the same file system.
	Copy(old, new URI, recursive bool) error
	// List lists files in a path.
	List(path URI, recursive bool) ([]Node, error)
	// Get gets a file or directory.
	Get(path URI) (Node, error)
	// MkDir creates a directory
	MkDir(path URI) (Node, error)
}

type Node struct {
	URI   URI
	IsDir bool
	// Filesystem??
	// Size??
	// LastModified??

}

func NewNode(uri URI, isDir bool) Node {
	return Node{
		URI:   uri,
		IsDir: isDir,
	}
}

/*
URI is a resource identifier. It is composed of a scheme and a path, separated by ":/"
such : [scheme]:/[path]
*/
type URI struct {
	// Scheme of the resource
	Scheme string
	// Path of the resource in the scheme
	Path string
	// Name of the resource
	Name string
}

func (u URI) String() string {
	return u.Scheme + "://" + u.Path
}

// NewURI creates a new URI from a scheme and a path and sets the name to the base of the path
func NewURI(s string, p string) URI {
	return URI{Scheme: s, Path: p, Name: path.Base(p)}
}

var ErrInvalidURI = errors.New("invalid URI")

var re = regexp.MustCompile(`^(?:(\w+):\/\/)?(.+)$`)

func ParseURI(uri string) (URI, error) {
	// It looks for an optional scheme followed by '://', and then captures the rest of the string
	matches := re.FindStringSubmatch(uri)
	if matches == nil || len(matches) != 3 {
		return NewURI("", uri), ErrInvalidURI
	}
	scheme, path := matches[1], matches[2]
	if scheme != "" && !ValidScheme(scheme) {
		return NewURI(scheme, path), ErrUnknownScheme
	}
	return NewURI(scheme, path), nil
}

const (
	LocalScheme     string = ""
	GCPBucketScheme string = "gs"
)

func ValidScheme(scheme string) bool {
	switch scheme {
	case LocalScheme, GCPBucketScheme:
		return true
	default:
		return false
	}
}

var schemes = map[string]FS{
	GCPBucketScheme: NewGCPBucketFS(),
	LocalScheme:     NewLocalFS(),
}

func SchemeFS(scheme string) FS {
	fs, ok := schemes[scheme]
	if !ok {
		return nil
	}
	return fs
}

var ErrUnknownScheme = errors.New("unknown scheme")
