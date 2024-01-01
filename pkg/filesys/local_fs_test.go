package filesys

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
)

var filesys = LocalFS{}

// TestCase is a test case structure, used to test the various methods of the FS interface
type TestCase struct {
	name   string // test case name
	src    URI    // source path
	dst    URI    // destination path
	rec    bool   // is recursive
	create bool   // should create tmp file/dir
	isDir  bool   // is dir
	err    error  // expected error
}

func TestDelete(t *testing.T) {
	testCases := []TestCase{
		{name: "Delete file", src: NewURI(LocalScheme, "*.txt"), rec: true, err: nil, create: true, isDir: false},
		{name: "Delete non-existent file", src: NewURI(LocalScheme, "non-existent-file"), rec: true, err: ErrNotFound, create: false, isDir: false},

		{name: "Delete dir recursively", src: NewURI(LocalScheme, "*"), rec: true, err: nil, create: true, isDir: true},
		{name: "Delete dir non-recursively", src: NewURI(LocalScheme, "*"), rec: false, err: ErrDirNotEmpty, create: true, isDir: true},
		{name: "Delete non-existent dir", src: NewURI(LocalScheme, "non-existent-dir"), rec: true, err: ErrNotFound, create: false, isDir: true},
	}
	for _, tc := range testCases {
		test := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if test.create {
				tmpPath := NewTmpDir(t, "", test.src.Path)
				test.src.Path = tmpPath // update src to tmpPath
			}
			err := filesys.Delete(test.src, test.rec)
			if err != nil {
				Assert(t, err, test.err)
			}
			PathMustNotExist(t, test.src.Path)
		})
	}
}

func TestList(t *testing.T) {
	testCases := []TestCase{
		{name: "List file", src: NewURI(LocalScheme, "*"), rec: true, err: nil, create: true, isDir: false},
		{name: "List non-existent file", src: NewURI(LocalScheme, "non-existent-file"), rec: true, err: ErrNotFound, create: false, isDir: false},

		{name: "List dir", src: NewURI(LocalScheme, "*"), rec: true, err: nil, create: true, isDir: true},
		{name: "List non-existent dir", src: NewURI(LocalScheme, "non-existent-dir"), rec: true, err: ErrNotFound, create: false, isDir: true},
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
			files, err := filesys.List(test.src, test.rec)
			if err != nil {
				Assert(t, err, test.err)
			}
			if test.create {
				if test.isDir && len(files) != 2 {
					t.Logf("Expected 2 files, got %v", len(files))
					t.Fail()
				} else if !test.isDir && len(files) != 1 {
					t.Logf("Expected 1 file, got %v", len(files))
					t.Fail()
				}
				RemoveTmp(t, test.src.Path)
			}
		})
	}
}

func TestCopy(t *testing.T) {
	testCases := []TestCase{
		{name: "Copy file", src: NewURI(LocalScheme, "*"), dst: NewURI(LocalScheme, "test"), rec: true, err: nil, create: true, isDir: false},
		{name: "Copy non-existent file", src: NewURI(LocalScheme, "non-existing-file"), dst: NewURI(LocalScheme, "test"), rec: true, err: ErrNotFound, create: false, isDir: false},
		{name: "Copy file to existing file", src: NewURI(LocalScheme, "*"), dst: NewURI(LocalScheme, "test.txt"), rec: true, err: ErrAlreadyExists, create: true, isDir: false},

		{name: "Copy dir", src: NewURI(LocalScheme, "*"), dst: NewURI(LocalScheme, "test"), rec: true, err: nil, create: true, isDir: true},
		{name: "Copy non-existent dir", src: NewURI(LocalScheme, "non-existing-dir"), dst: NewURI(LocalScheme, "test"), rec: true, err: ErrNotFound, create: false, isDir: true},
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
			err := filesys.Copy(test.src, test.dst, test.rec)
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
		{name: "Get file recursive", src: NewURI(LocalScheme, "test1.txt"), rec: true, err: nil, create: true, isDir: false},
		{name: "Get file non-recursive", src: NewURI(LocalScheme, "test1.txt"), rec: false, err: nil, create: true, isDir: false},
		{name: "Get non-existent file", src: NewURI(LocalScheme, "test1.txt"), rec: true, err: ErrNotFound, create: false, isDir: false},

		{name: "Get dir recursive", src: NewURI(LocalScheme, "test1"), rec: true, err: nil, create: true, isDir: true},
		{name: "Get dir non-recursive", src: NewURI(LocalScheme, "test1"), rec: false, err: nil, create: true, isDir: true},
		{name: "Get non-existent dir", src: NewURI(LocalScheme, "test1"), rec: true, err: ErrNotFound, create: false, isDir: true},
	}
	for _, tc := range testCases {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.create {
				tmpPath := NewTmpDirOrFile(t, test.src.Path, test.isDir)
				test.src = tmpPath // update src to tmpPath
			}
			node, err := filesys.Get(test.src)
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

// NewTmpDirOrFile creates a tmp dir or file depending on isDir and returns the path
func NewTmpDirOrFile(t *testing.T, dir string, isDir bool) URI {
	t.Helper()
	var tmpPath string
	if isDir {
		tmpPath = NewTmpDir(t, "", dir)
	} else {
		tmpPath = NewTmpFile(t, "", dir)
	}
	return NewURI(LocalScheme, tmpPath)
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
	if _, err := os.Stat(dir); errors.Is(err, ErrNotFound) {
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
