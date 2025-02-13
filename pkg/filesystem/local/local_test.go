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
	t.Run("file operations", testListFile)
	t.Run("directory operations", testListDirectory)
	t.Run("recursive operations", testListRecursive)
	t.Run("error cases", testListErrors)
}

// findNode recursively searches for a node with the given name in the tree
func findNode(tree filesystem.NodeTree, name string) *filesystem.NodeTree {
	if filepath.Base(tree.Node.Path) == name {
		return &tree
	}
	for _, child := range tree.Children {
		if found := findNode(child, name); found != nil {
			return found
		}
	}
	return nil
}

func testListFile(t *testing.T) {
	t.Parallel()

	// Setup
	tmpDir := NewTmpDir(t, "", "test-list-file-*")
	defer RemoveTmp(t, tmpDir)

	// Create file with exact name (not using temp file pattern)
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test listing a single file
	ctx := context.Background()
	result, err := localFS.List(ctx, uri.URI{
		Scheme: uri.LocalScheme,
		Path:   filePath,
	}, false)

	// Assertions
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	Assert(t, result.Node.IsDir, false)
	Assert(t, len(result.Children), 0)
	Assert(t, filepath.Base(result.Node.Path), "test.txt")
}

func testListDirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string) // setup function returns expected files
		expected []string
	}{
		{
			name: "empty directory",
			setup: func(t *testing.T, dir string) {
				// No files to create
			},
			expected: []string{},
		},
		{
			name: "directory with files",
			setup: func(t *testing.T, dir string) {
				createFileWithName(t, dir, "file1.txt")
				createFileWithName(t, dir, "file2.txt")
			},
			expected: []string{"file1.txt", "file2.txt"},
		},
		{
			name: "directory with subdirectory (non-recursive)",
			setup: func(t *testing.T, dir string) {
				createFileWithName(t, dir, "file1.txt")
				createDirWithName(t, dir, "subdir")
			},
			expected: []string{"file1.txt", "subdir"},
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			tmpDir := NewTmpDir(t, "", "test-list-dir-*")
			defer RemoveTmp(t, tmpDir)
			test.setup(t, tmpDir)

			// Test
			ctx := context.Background()
			result, err := localFS.List(ctx, uri.URI{
				Scheme: uri.LocalScheme,
				Path:   tmpDir,
			}, false)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			Assert(t, result.Node.IsDir, true)
			Assert(t, len(result.Children), len(test.expected))

			// Verify all expected files are present
			found := make(map[string]bool)
			for _, child := range result.Children {
				found[filepath.Base(child.Node.Path)] = true
			}
			for _, expectedFile := range test.expected {
				if !found[expectedFile] {
					t.Errorf("expected file %s not found in results", expectedFile)
				}
			}
		})
	}
}

// DirectoryStructure represents the expected contents of directories
type DirectoryStructure map[string][]string

// verifyDir recursively verifies that the tree matches the expected structure
func verifyDir(t *testing.T, tree filesystem.NodeTree, expectedStructure DirectoryStructure) {
	t.Helper()
	path := filepath.Base(tree.Node.Path)
	expectedContents := expectedStructure[path]
	foundContents := make(map[string]bool)

	for _, child := range tree.Children {
		foundContents[filepath.Base(child.Node.Path)] = true
		if child.Node.IsDir {
			verifyDir(t, child, expectedStructure)
		}
	}

	// Verify all expected contents are present
	for _, expected := range expectedContents {
		if !foundContents[expected] {
			t.Errorf("in directory %q: expected item %q not found", path, expected)
		}
	}

	// Verify no unexpected contents
	for found := range foundContents {
		isExpected := false
		for _, expected := range expectedContents {
			if found == expected {
				isExpected = true
				break
			}
		}
		if !isExpected {
			t.Errorf("in directory %q: unexpected item %q found", path, found)
		}
	}
}

func testListRecursive(t *testing.T) {
	t.Parallel()

	// Define the expected structure
	expectedStructure := DirectoryStructure{
		// root dir name will be set after creation
		"":          {"root1.txt", "root2.txt", "subdir1", "subdir2"},
		"subdir1":   {"sub1_file1.txt", "subsubdir"},
		"subdir2":   {"sub2_file1.txt"},
		"subsubdir": {"deep_file.txt"},
	}

	// Setup complex directory structure
	tmpDir := NewTmpDir(t, "", "test-list-recursive-*")
	defer RemoveTmp(t, tmpDir)

	createFileWithName(t, tmpDir, "root1.txt")
	createFileWithName(t, tmpDir, "root2.txt")
	subDir1 := createDirWithName(t, tmpDir, "subdir1")
	subDir2 := createDirWithName(t, tmpDir, "subdir2")
	createFileWithName(t, subDir1, "sub1_file1.txt")
	createFileWithName(t, subDir2, "sub2_file1.txt")
	subSubDir := createDirWithName(t, subDir1, "subsubdir")
	createFileWithName(t, subSubDir, "deep_file.txt")

	// Update root directory name in expected structure
	rootDirName := filepath.Base(tmpDir)
	expectedStructure[rootDirName] = expectedStructure[""]
	delete(expectedStructure, "")

	// Test
	ctx := context.Background()
	result, err := localFS.List(ctx, uri.URI{
		Scheme: uri.LocalScheme,
		Path:   tmpDir,
	}, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the entire structure
	verifyDir(t, result, expectedStructure)

	// Verify specific paths exist and have correct properties
	subdir1Node := findNode(result, "subdir1")
	if subdir1Node == nil {
		t.Fatal("subdir1 not found in results")
	}
	Assert(t, subdir1Node.Node.IsDir, true)
	Assert(t, len(subdir1Node.Children), 2) // sub1_file1.txt and subsubdir

	// Verify subsubdir structure
	subsubdirNode := findNode(result, "subsubdir")
	if subsubdirNode == nil {
		t.Fatal("subsubdir not found in results")
	}
	Assert(t, subsubdirNode.Node.IsDir, true)
	Assert(t, len(subsubdirNode.Children), 1) // deep_file.txt
}

func testListErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		expectedErr error
	}{
		{
			name:        "non-existent path",
			path:        "/path/that/does/not/exist",
			expectedErr: filesystem.ErrNotFound,
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			_, err := localFS.List(ctx, uri.URI{
				Scheme: uri.LocalScheme,
				Path:   test.path,
			}, false)

			Assert(t, err, test.expectedErr)
		})
	}
}

func TestCopy(t *testing.T) {
	testCases := []TestCase{
		{name: "Copy file",
			src: uri.URI{Scheme: uri.LocalScheme, Path: "*"},
			dst: uri.URI{Scheme: uri.LocalScheme, Path: "test"},
			rec: true, err: nil,
			create: true, isDir: false,
		},
		{name: "Copy non-existent file",
			src: uri.URI{Scheme: uri.LocalScheme, Path: "non-existing-file"},
			dst: uri.URI{Scheme: uri.LocalScheme, Path: "test"},
			rec: true, err: filesystem.ErrNotFound,
			create: false, isDir: false,
		},
		{name: "Copy file to existing file",
			src: uri.URI{Scheme: uri.LocalScheme, Path: "*"},
			dst: uri.URI{Scheme: uri.LocalScheme, Path: "test.txt"},
			rec: true, err: filesystem.ErrAlreadyExists,
			create: true, isDir: false,
		},
		{name: "Copy dir",
			src: uri.URI{Scheme: uri.LocalScheme, Path: "*"},
			dst: uri.URI{Scheme: uri.LocalScheme, Path: "test"},
			rec: true, err: nil,
			create: true, isDir: true,
		},
		{name: "Copy non-existent dir",
			src: uri.URI{Scheme: uri.LocalScheme, Path: "non-existing-dir"},
			dst: uri.URI{Scheme: uri.LocalScheme, Path: "test"},
			rec: true, err: filesystem.ErrNotFound,
			create: false, isDir: true,
		},
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
		{name: "Get file non-recursive", src: uri.URI{Scheme: uri.LocalScheme, Path: "test1.txt"},
			rec: false, err: nil, create: true, isDir: false},
		{name: "Get non-existent file", src: uri.URI{Scheme: uri.LocalScheme, Path: "test1.txt"},
			rec: true, err: filesystem.ErrNotFound, create: false, isDir: false},
		{name: "Get dir recursive", src: uri.URI{Scheme: uri.LocalScheme, Path: "test1"},
			rec: true, err: nil, create: true, isDir: true},
		{name: "Get dir non-recursive", src: uri.URI{Scheme: uri.LocalScheme, Path: "test1"},
			rec: false, err: nil, create: true, isDir: true},
		{name: "Get non-existent dir", src: uri.URI{Scheme: uri.LocalScheme, Path: "test1"},
			rec: true, err: filesystem.ErrNotFound, create: false, isDir: true},
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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			// Create absolute path under our temp directory
			absPath := filepath.Join(tmpRoot, tc.src.Path)
			tc.src.Path = absPath

			if tc.err == filesystem.ErrAlreadyExists {
				_, err := localFS.MkDir(ctx, tc.src)
				if err != nil {
					t.Fatalf("Failed to create directory for already exists test: %v", err)
				}
			}

			node, err := localFS.MkDir(ctx, tc.src)
			Assert(t, err, tc.err)

			if err == nil {
				PathMustExist(t, tc.src.Path)
				Assert(t, node.URI.Path, tc.src.Path)
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

func createFileWithName(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create file %s: %v", name, err)
	}
	return path
}

func createDirWithName(t *testing.T, parent, name string) string {
	t.Helper()
	path := filepath.Join(parent, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", name, err)
	}
	return path
}
