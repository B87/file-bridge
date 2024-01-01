package filesys

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// GCPBucketFS is a FileSystem implementation that uses a GCP bucket.
type GCPBucketFS struct {
	ctx    context.Context
	client *storage.Client
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
	bucket, object := SplitGCPPath(name.Path)
	wc := fs.client.Bucket(bucket).Object(object).NewWriter(fs.ctx)
	return wc, nil
}

func (fs *GCPBucketFS) Reader(name URI) (io.ReadCloser, error) {
	bucket, object := SplitGCPPath(name.Path)
	rc, err := fs.client.Bucket(bucket).Object(object).NewReader(fs.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", name.Path, err)
	}
	return rc, nil
}

func (fs *GCPBucketFS) Delete(name URI, recursive bool) error {
	return nil
}
func (fs *GCPBucketFS) Move(oldName, newName URI, recursive bool) error {
	return nil
}

// Copy copies a file from one path to another inside the same GS filesystem.
// Use manager Copy for cross filesystem copy.
func (fs *GCPBucketFS) Copy(src, dst URI, recursive bool) error {
	srcBucket, srcObject := SplitGCPPath(src.Path)
	srcObj := fs.client.Bucket(srcBucket).Object(srcObject)
	dstBucket, dstObject := SplitGCPPath(dst.Path)
	dstObj := fs.client.Bucket(dstBucket).Object(dstObject)
	_, err := dstObj.CopierFrom(srcObj).Run(fs.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (fs *GCPBucketFS) List(dir URI, recursive bool) ([]Node, error) {
	var files []Node
	bucket, object := SplitGCPPath(dir.Path)
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
			files = append(files, NewNode(NewURI(dir.Scheme, attrs.Name), false))
		}

	}
	return files, nil
}

func (fs *GCPBucketFS) Get(uri URI) (Node, error) {
	bucket, object := SplitGCPPath(uri.Path)
	object = strings.TrimSuffix(object, "/")
	gsObj := fs.client.Bucket(bucket).Object(object)
	_, err := gsObj.Attrs(fs.ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return NewNode(uri, false), ErrNotFound
		}
		return Node{}, err
	}
	return NewNode(uri, false), nil
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

func (fs *GCPBucketFS) IsGSDir(uri URI) (bool, error) {

	bucketName, objectName := SplitGCPPath(uri.Path)
	objectName = strings.TrimSuffix(objectName, "/") // Ensure the URI does not end with a slash

	it := fs.client.Bucket(bucketName).Objects(fs.ctx, &storage.Query{Prefix: objectName})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return false, err
		}

		if attrs.Name == objectName {
			return false, nil // Exact match found, it's an object
		}
	}
	return true, nil // No exact match found, likely a directory
}

func SplitGCPPath(path string) (bucket string, object string) {
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
