package filesys

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
)

type StorageClient interface {
	Bucket(name string) *storage.BucketHandle
	Close() error
}

// GCPBucketFS is a FileSystem implementation that uses a GCP bucket.
type GCPBucketFS struct {
	ctx    context.Context
	client StorageClient
}

func NewGCPBucketFS() *GCPBucketFS {
	return &GCPBucketFS{}
}

func (fs *GCPBucketFS) Connect() error {
	fs.ctx = context.Background()
	client, err := storage.NewClient(fs.ctx)
	if err != nil {
		return err
	}
	fs.client = client
	return nil
}
func (fs *GCPBucketFS) Disconnect() error {
	return fs.client.Close()
}

func (fs *GCPBucketFS) Writer(name URI) (io.WriteCloser, error) {
	bucket, object := splitGCPPath(name.Path)
	wc := fs.client.Bucket(bucket).Object(object).NewWriter(fs.ctx)
	return wc, nil
}

func (fs *GCPBucketFS) Reader(name URI) (io.ReadCloser, error) {
	bucket, object := splitGCPPath(name.Path)
	rc, err := fs.client.Bucket(bucket).Object(object).NewReader(fs.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", name.Path, err)
	}
	return rc, nil
}

func (fs *GCPBucketFS) Delete(name URI, recursive bool) error {
	bucket, object := splitGCPPath(name.Path)
	it := fs.client.Bucket(bucket).Objects(fs.ctx, &storage.Query{Prefix: object})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		err = fs.client.Bucket(bucket).Object(attrs.Name).Delete(fs.ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Copy copies a file from one path to another inside the same GS filesystem.
// Use manager Copy for cross filesystem copy.
func (fs *GCPBucketFS) Copy(src, dst URI, recursive bool) error {
	srcBucket, srcObject := splitGCPPath(src.Path)
	srcObj := fs.client.Bucket(srcBucket).Object(srcObject)
	dstBucket, dstObject := splitGCPPath(dst.Path)
	dstObj := fs.client.Bucket(dstBucket).Object(dstObject)
	_, err := dstObj.CopierFrom(srcObj).Run(fs.ctx)
	if err != nil {
		return err
	}
	return nil
}

// List lists files and folders in a path.
func (fs *GCPBucketFS) List(dir URI, recursive bool) ([]Node, error) {
	var files []Node
	bucket, object := splitGCPPath(dir.Path)
	query := storage.Query{Prefix: object}
	if !recursive {
		query.Delimiter = "/"
	}
	it := fs.client.Bucket(bucket).Objects(fs.ctx, &query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		if attrs.Prefix != "" {
			// This is a 'folder'
			files = append(files, NewNode(NewURI(dir.Scheme, attrs.Prefix), true))
		} else {
			// This is a file
			files = append(files, NewNode(NewURI(dir.Scheme, path.Join(attrs.Bucket, attrs.Name)), false))
		}
	}
	return files, nil
}

func (fs *GCPBucketFS) Get(uri URI) (Node, error) {
	bucket, object := splitGCPPath(uri.Path)
	object = strings.TrimSuffix(object, "/")

	it := fs.client.Bucket(bucket).Objects(
		fs.ctx, &storage.Query{Prefix: object})

	found := false
	isDir := false

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return Node{}, err
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
		return NewNode(uri, false), ErrNotFound
	}
	return NewNode(uri, isDir), nil
}

func (fs *GCPBucketFS) MkDir(path URI) (Node, error) {
	// Make sure path ends with a slash
	if !strings.HasSuffix(path.Path, "/") {
		path.Path = path.Path + "/"
	}
	w, err := fs.Writer(path)
	if err != nil {
		return Node{}, err
	}
	defer w.Close()
	if _, err := io.Copy(w, &bytes.Buffer{}); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	return NewNode(path, true), nil
}

func splitGCPPath(path string) (bucket string, object string) {
	tree := strings.Split(path, "/")
	if len(tree) < 1 {
		return "", ""
	}
	bucket = tree[0]
	if len(tree) < 2 {
		return bucket, ""
	}
	object = strings.Join(tree[1:], "/")
	return bucket, object
}
