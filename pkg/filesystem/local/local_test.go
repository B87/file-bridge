package local

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/B87/file-bridge/pkg/filesystem"
	"github.com/B87/file-bridge/pkg/uri"
)

var localFS = LocalFS{}

// TestCase is a test case structure, used to test the various methods of the FS interface
type TestCase struct {
	name   string  // test case name
	src    uri.URI // source path
	dst    uri.URI // destination path
	create bool    // should create tmp file/dir
	isDir  bool    // is dir
	rec    bool    // is recursive
	err    error   // expected error
}

func TestDelete(t *testing.T) {
	testCases := []TestCase{
		{name: "Delete file", src: uri.URI{Scheme: uri.LocalScheme, Path: "*.txt"}, err: nil, create: true, isDir: false},
		{name: "Delete non-existent file", src: uri.URI{Scheme: uri.LocalScheme, Path: "non-existent-file"}, err: filesystem.ErrNotFound, create: false, isDir: false},

		{name: "Delete dir recursively", src: uri.URI{Scheme: uri.LocalScheme, Path: "*"}, err: nil, create: true, isDir: true},
		{name: "Delete dir non-recursively", src: uri.URI{Scheme: uri.LocalScheme, Path: "*"}, err: filesystem.ErrDirNotEmpty, create: true, isDir: true},
		{name: "Delete non-existent dir", src: uri.URI{Scheme: uri.LocalScheme, Path: "non-existent-dir"}, err: filesystem.ErrNotFound, create: false, isDir: true},
	}
	for _, tc := range testCases {
		test := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if test.create {
				tmpPath := NewTmpDir(t, "", test.src.Path)
				test.src.Path = tmpPath // update src to tmpPath
			}
			ctx := context.Background()
			err := localFS.Delete(ctx, test.src)
			if err != nil {
				Assert(t, err, test.err)
			}
			PathMustNotExist(t, test.src.Path)
		})
	}
}

func TestList(t *testing.T) {
	testCases := []TestCase{
		{name: "List file", src: uri.URI{Scheme: uri.LocalScheme, Path: "*"}, rec: true, err: nil, create: true, isDir: false},
		{name: "List non-existent file", src: uri.URI{Scheme: uri.LocalScheme, Path: "non-existent-file"}, rec: true, err: filesystem.ErrNotFound, create: false, isDir: false},

		{name: "List dir", src: uri.URI{Scheme: uri.LocalScheme, Path: "*"}, rec: true, err: nil, create: true, isDir: true},
		{name: "List non-existent dir", src: uri.URI{Scheme: uri.LocalScheme, Path: "non-existent-dir"}, rec: true, err: filesystem.ErrNotFound, create: false, isDir: true},
	}
	for _, tc := range testCases {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.create {
				tmpPath := NewTmpDir(t, "", test.src.Path)
				file1 := NewTmpFile(t, tmpPath, "test1.txt")
				NewTmpFile(t, tmpPath, "test2.txt")
				if test.isDir {
					test.src.Path = tmpPath // update src to tmp folder
				} else {
					test.src.Path = file1 // update src to tmp file 1
				}
			}
			ctx := context.Background()
			files, err := localFS.List(ctx, test.src, test.rec)
			if err != nil {
				Assert(t, err, test.err)
			}
			if test.create {
				if test.isDir && len(files.Children) != 2 {
					t.Logf("Expected 2 files, got %v", len(files.Children))
					t.Fail()
				} else if !test.isDir && len(files.Children) != 0 {
					t.Logf("Expected 1 file, got %v", len(files.Children))
					t.Fail()
				}
				RemoveTmp(t, test.src.Path)
			}
		})
	}
}

func TestCopy(t *testing.T) {
	testCases := []TestCase{
		{name: "Copy file",
			src:    uri.URI{Scheme: uri.LocalScheme, Path: "*"},
			dst:    uri.URI{Scheme: uri.LocalScheme, Path: "test"},
			rec:    true,
			err:    nil,
			create: true,
			isDir:  false,
		},
		{name: "Copy non-existent file",
			src:    uri.URI{Scheme: uri.LocalScheme, Path: "non-existing-file"},
			dst:    uri.URI{Scheme: uri.LocalScheme, Path: "test"},
			rec:    true,
			err:    filesystem.ErrNotFound,
			create: false,
			isDir:  false,
		},
		{name: "Copy file to existing file",
			src: uri.URI{Scheme: uri.LocalScheme, Path: "*"},
			dst: uri.URI{Scheme: uri.LocalScheme, Path: "test.txt"},
			rec: true, err: filesystem.ErrAlreadyExists,
			create: true, isDir: false,
		},
		{name: "Copy dir",
			src:    uri.URI{Scheme: uri.LocalScheme, Path: "*"},
			dst:    uri.URI{Scheme: uri.LocalScheme, Path: "test"},
			rec:    true,
			err:    nil,
			create: true,
			isDir:  true,
		},
		{name: "Copy non-existent dir",
			src: uri.URI{Scheme: uri.LocalScheme, Path: "non-existing-dir"}, dst: uri.URI{Scheme: uri.LocalScheme, Path: "test"}, rec: true, err: filesystem.ErrNotFound, create: false, isDir: true},
	}
	for _, tc := range testCases {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.create {
				tmpPath := NewTmpDirOrFile(t, test.src.Path, test.isDir)
				test.src = tmpPath // update src to tmpPath
				dstPath := NewTmpDir(t, "", test.dst.Path)
				test.dst.Path = dstPath
			}
			ctx := context.Background()
			err := localFS.Copy(ctx, test.src, test.dst)
			if err != nil {
				Assert(t, err, test.err)
			}
			if test.create {
				PathMustExist(t, test.src.Path)
				RemoveTmp(t, test.src.Path)
				PathMustExist(t, test.dst.Path)
				RemoveTmp(t, test.dst.Path)
			}
		})
	}
}

func TestGet(t *testing.T) {
	testCases := []TestCase{
		{name: "Get file recursive", src: uri.URI{Scheme: uri.LocalScheme, Path: "test1.txt"},
			rec: true, err: nil, create: true, isDir: false},
		{name: "Get file non-recursive", src: uri.URI{Scheme: uri.LocalScheme, Path: "test1.txt"}, rec: false, err: nil, create: true, isDir: false},
		{name: "Get non-existent file", src: uri.URI{Scheme: uri.LocalScheme, Path: "test1.txt"}, rec: true, err: filesystem.ErrNotFound, create: false, isDir: false},

		{name: "Get dir recursive", src: uri.URI{Scheme: uri.LocalScheme, Path: "test1"}, rec: true, err: nil, create: true, isDir: true},
		{name: "Get dir non-recursive", src: uri.URI{Scheme: uri.LocalScheme, Path: "test1"}, rec: false, err: nil, create: true, isDir: true},
		{name: "Get non-existent dir", src: uri.URI{Scheme: uri.LocalScheme, Path: "test1"}, rec: true, err: filesystem.ErrNotFound, create: false, isDir: true},
	}
	for _, tc := range testCases {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.create {
				tmpPath := NewTmpDirOrFile(t, test.src.Path, test.isDir)
				test.src = tmpPath // update src to tmpPath
			}
			ctx := context.Background()
			node, err := localFS.Get(ctx, test.src)
			if err != nil {
				Assert(t, err, test.err)
			}
			if test.create {
				splitPath := strings.Split(test.src.Path, "/")
				Assert(t, node.URI.Name, splitPath[len(splitPath)-1])
				RemoveTmp(t, test.src.Path)
			}
		})
	}
}

func TestMkDir(t *testing.T) {
	// Create a temporary root directory for all tests
	tmpRoot := NewTmpDir(t, "", "mkdirtest")
	defer RemoveTmp(t, tmpRoot)

	testCases := []TestCase{
		{name: "MkDir", src: uri.URI{Scheme: uri.LocalScheme, Path: "test"}, err: nil},
		{name: "MkDir already exists", src: uri.URI{Scheme: uri.LocalScheme, Path: "test"}, err: filesystem.ErrAlreadyExists},
		{name: "MkDir recursive", src: uri.URI{Scheme: uri.LocalScheme, Path: "test/rec"}, err: nil},
	}
	for _, tc := range testCases {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			// Create absolute path under our temp directory
			absPath := filepath.Join(tmpRoot, test.src.Path)
			test.src.Path = absPath

			if test.err == filesystem.ErrAlreadyExists {
				// First create the directory to test the "already exists" case
				_, err := localFS.MkDir(ctx, test.src)
				if err != nil {
					t.Fatalf("Failed to create directory for already exists test: %v", err)
				}
			}

			node, err := localFS.MkDir(ctx, test.src)
			Assert(t, err, test.err)

			if err == nil {
				PathMustExist(t, test.src.Path)
				Assert(t, node.URI.Path, test.src.Path)
			}

		})
	}
}

// NewTmpDirOrFile creates a tmp dir or file depending on isDir and returns the path
func NewTmpDirOrFile(t *testing.T, dir string, isDir bool) uri.URI {
	t.Helper()
	var tmpPath string
	if isDir {
		tmpPath = NewTmpDir(t, "", dir)
	} else {
		tmpPath = NewTmpFile(t, "", dir)
	}
	// Get the base name from the path
	name := filepath.Base(tmpPath)
	return uri.URI{
		Scheme: uri.LocalScheme,
		Path:   tmpPath,
		Name:   name,
	}
}

// Assert compares got and want and logs a message if they are not equal
func Assert(t *testing.T, got, want interface{}) {
	t.Helper()
	switch got.(type) {
	case error:
		want, ok := want.(error)
		if ok && !errors.Is(got.(error), want) {
			t.Logf("got: %v, expected: %v", got, want)
			t.Fail()
		}
	default:
		if got != want {
			t.Logf("got: %v, want: %v", got, want)
			t.Fail()
		}
	}
}

func PathMustExist(t *testing.T, dir string) {
	t.Helper()
	if _, err := os.Stat(dir); errors.Is(err, filesystem.ErrNotFound) {
		t.Logf("Directory %s does not exist", dir)
		t.Fail()
	}
}

func PathMustNotExist(t *testing.T, dir string) {
	t.Helper()
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return // ok
	} else if err != nil {
		t.Logf("Error checking if directory %s exists: %v", dir, err)
		t.Fail()
	} else {
		t.Logf("Directory %s exists", dir)
		t.Fail()
	}
}

func NewTmpDir(t *testing.T, root, name string) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp(root, name)
	if err != nil {
		t.Logf("Error creating tmp dir %s: %v", tmpDir, err)
		t.Fail()
	}
	PathMustExist(t, tmpDir)
	return tmpDir
}

func NewTmpFile(t *testing.T, root, name string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp(root, name)
	if err != nil {
		t.Logf("Error creating tmp file %s: %v", tmpFile.Name(), err)
		t.Fail()
	}
	PathMustExist(t, tmpFile.Name())
	return tmpFile.Name()
}

func RemoveTmp(t *testing.T, dir string) {
	t.Helper()
	err := os.RemoveAll(dir)
	if err != nil {
		t.Logf("Error removing tmp dir %s: %v", dir, err)
		t.Fail()
	}
	PathMustNotExist(t, dir)
}
