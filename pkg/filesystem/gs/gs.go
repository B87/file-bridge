package gs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"path"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"

	"github.com/B87/file-bridge/pkg/filesystem"
	"github.com/B87/file-bridge/pkg/uri"
)

type StorageClient interface {
	Bucket(name string) *storage.BucketHandle
	Close() error
}

// GCPBucketFS is a FileSystem implementation that uses a GCP bucket.
type GCPBucketFS struct {
	client StorageClient
}

// NewGCPBucketFS creates a new GCPBucketFS.
func NewGCPBucketFS(ctx context.Context) *GCPBucketFS {
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create GCP client: %v", err)
	}
	return &GCPBucketFS{client: client}
}

// Writer returns a WriteCloser for the given URI.
func (fs *GCPBucketFS) Writer(ctx context.Context, uri uri.URI) (io.WriteCloser, error) {
	bucket, object := filesystem.SplitPath(uri.Path)
	wc := fs.client.Bucket(bucket).Object(object).NewWriter(ctx)
	return wc, nil
}

// Reader returns a ReadCloser for the given URI.
func (fs *GCPBucketFS) Reader(ctx context.Context, uri uri.URI) (io.ReadCloser, error) {
	bucket, object := filesystem.SplitPath(uri.Path)
	rc, err := fs.client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", uri.Path, err)
	}
	return rc, nil
}

// Delete deletes the file or directory at the given URI.
func (fs *GCPBucketFS) Delete(ctx context.Context, uri uri.URI) error {
	bucket, object := filesystem.SplitPath(uri.Path)
	it := fs.client.Bucket(bucket).Objects(ctx, &storage.Query{Prefix: object})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		err = fs.client.Bucket(bucket).Object(attrs.Name).Delete(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Copy copies a file from one path to another inside the same GS filesystem.
// Use manager Copy for cross filesystem copy.
func (fs *GCPBucketFS) Copy(ctx context.Context, src, dst uri.URI) error {
	srcBucket, srcObject := filesystem.SplitPath(src.Path)
	srcObj := fs.client.Bucket(srcBucket).Object(srcObject)
	dstBucket, dstObject := filesystem.SplitPath(dst.Path)
	dstObj := fs.client.Bucket(dstBucket).Object(dstObject)
	_, err := dstObj.CopierFrom(srcObj).Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

// List lists files and folders in a path (dir).
func (fs *GCPBucketFS) List(ctx context.Context, dir uri.URI, recursive bool) (filesystem.NodeTree, error) {
	bucketName, objectPrefix := filesystem.SplitPath(dir.Path)
	// Ensure consistent prefix handling (no trailing slash, but we'll add it if non-empty)
	objectPrefix = strings.TrimSuffix(objectPrefix, "/")
	if objectPrefix != "" {
		objectPrefix += "/"
	}

	query := &storage.Query{Prefix: objectPrefix}
	if !recursive {
		query.Delimiter = "/"
	}

	// Our root node. This corresponds to the top-level "directory" the user requested.
	// We'll store it in nodeMap under the empty string "" to mean "this is the root relative path."
	root := filesystem.NodeTree{
		Node: filesystem.NewNode(dir, true),
	}

	// nodeMap key = relative path from objectPrefix, e.g. "", "subdir", "subdir/file.txt"
	// The value = pointer to the NodeTree for that path
	nodeMap := make(map[string]*filesystem.NodeTree)
	nodeMap[""] = &root

	// Keep track of everything we've seen to avoid duplicates
	seenPaths := make(map[string]bool)

	it := fs.client.Bucket(bucketName).Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return filesystem.NodeTree{}, fmt.Errorf("failed to iterate: %w", err)
		}

		// Check if it's a "directory" placeholder in non-recursive mode:
		if !recursive && attrs.Prefix != "" {
			// This is a subdirectory (GCS calls it a 'prefix')
			dirRelPath := strings.TrimPrefix(attrs.Prefix, objectPrefix) // relative to the original prefix
			dirRelPath = strings.TrimSuffix(dirRelPath, "/")             // remove trailing slash
			if dirRelPath == "" {
				// sometimes GCS gives back the exact prefix in attrs.Prefix, skip if empty
				continue
			}
			if seenPaths[dirRelPath] {
				continue
			}
			seenPaths[dirRelPath] = true

			// Create NodeTree for this directory
			dirNode := &filesystem.NodeTree{
				Node: filesystem.NewNode(
					uri.URI{
						Scheme: "gs", // or whatever your scheme
						Path:   path.Join(bucketName, dirRelPath),
					},
					true, // isDir
				),
			}
			// Insert into the map so children can attach later
			nodeMap[dirRelPath] = dirNode

			// Add to parent's children
			parentRelPath := path.Dir(dirRelPath)
			if parentRelPath == "." {
				parentRelPath = "" // means root
			}
			parent, ok := nodeMap[parentRelPath]
			if !ok {
				// If we never saw the parent, create it on the fly or skip
				parent = &filesystem.NodeTree{
					Node: filesystem.NewNode(uri.URI{
						Scheme: "gs",
						Path:   path.Join(bucketName, parentRelPath),
					}, true),
				}
				nodeMap[parentRelPath] = parent
			}
			parent.Children = append(parent.Children, *dirNode)
			continue
		}

		// If we reach here, we have a regular object (file) or possibly a folder ending with slash in GCS
		if attrs.Name == "" {
			// This can happen for the "prefix" itself. Skip empty.
			continue
		}

		// A true GCS object name (file)
		relPath := strings.TrimPrefix(attrs.Name, objectPrefix) // strip the original prefix
		if relPath == "" {
			// It's exactly the prefix item, skip
			continue
		}
		// Chop off trailing slash if the object name has one
		relPath = strings.TrimSuffix(relPath, "/")

		if seenPaths[relPath] {
			continue
		}
		seenPaths[relPath] = true

		// Build up intermediate directories if needed
		parts := strings.Split(relPath, "/")
		for i := 1; i < len(parts); i++ {
			subDir := path.Join(parts[:i]...)
			// If we haven't seen this subdirectory, create it
			if !seenPaths[subDir] {
				seenPaths[subDir] = true
				subDirNode := &filesystem.NodeTree{
					Node: filesystem.NewNode(
						uri.URI{
							Scheme: "gs",
							Path:   path.Join(bucketName, subDir),
						},
						true),
				}
				nodeMap[subDir] = subDirNode

				// Attach it to its parent
				parentRel := path.Dir(subDir)
				if parentRel == "." {
					parentRel = ""
				}
				parent, ok := nodeMap[parentRel]
				if !ok {
					// create parent if it doesn't exist
					parent = &filesystem.NodeTree{
						Node: filesystem.NewNode(uri.URI{
							Scheme: "gs",
							Path:   path.Join(bucketName, parentRel),
						}, true),
					}
					nodeMap[parentRel] = parent
				}
				parent.Children = append(parent.Children, *subDirNode)
			}
		}

		// Finally, create the file node
		fileNode := &filesystem.NodeTree{
			Node: filesystem.NewNode(
				uri.URI{
					Scheme: "gs",
					Path:   path.Join(bucketName, relPath),
				},
				false), // file, not dir
		}
		// If you want to propagate size, modTime, etc. from attrs, do it here:
		fileNode.Node.Size = attrs.Size
		fileNode.Node.ModTime = attrs.Updated
		// Or set Mode, etc.

		nodeMap[relPath] = fileNode

		// Attach it to parent
		parentRelPath := path.Dir(relPath)
		if parentRelPath == "." {
			parentRelPath = ""
		}
		parent, ok := nodeMap[parentRelPath]
		if !ok {
			// create the parent dir on the fly if we never saw it
			parent = &filesystem.NodeTree{
				Node: filesystem.NewNode(uri.URI{
					Scheme: "gs",
					Path:   path.Join(bucketName, parentRelPath),
				}, true),
			}
			nodeMap[parentRelPath] = parent
		}
		parent.Children = append(parent.Children, *fileNode)
	}

	return root, nil
}

// Get returns the node at the given URI.
func (fs *GCPBucketFS) Get(ctx context.Context, uri uri.URI) (filesystem.Node, error) {
	bucket, object := filesystem.SplitPath(uri.Path)
	object = strings.TrimSuffix(object, "/")

	it := fs.client.Bucket(bucket).Objects(ctx, &storage.Query{Prefix: object})

	found := false
	isDir := false

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return filesystem.Node{}, err
		}
		if attrs.Name == object {
			found = true // Exact match, it's a file
			break
		} else if strings.HasPrefix(attrs.Name, object+"/") {
			found = true
			isDir = true // Objects found with the prefix, it's a directory
			break
		}
	}
	if !found {
		return filesystem.Node{}, filesystem.ErrNotFound
	}
	return filesystem.NewNode(uri, isDir), nil
}

// MkDir creates a new directory in the GCP bucket.
// It creates an empty file with the name of the directory.
func (fs *GCPBucketFS) MkDir(ctx context.Context, path uri.URI) (filesystem.Node, error) {
	// Make sure path ends with a slash
	path.Path = filesystem.SuffixSlash(path.Path)
	w, err := fs.Writer(ctx, path)
	if err != nil {
		return filesystem.Node{}, err
	}
	defer w.Close()
	if _, err := io.Copy(w, &bytes.Buffer{}); err != nil {
		return filesystem.Node{}, err
	}
	return filesystem.NewNode(path, true), nil
}
