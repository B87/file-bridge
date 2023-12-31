package filesys

import (
	"context"
	"io"
	"log"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// GCPBucketFS is a FileSystem implementation that uses a GCP bucket.
type GCPBucketFS struct{}

func NewGCPBucketFS() *GCPBucketFS {
	return &GCPBucketFS{}
}

func (GCPBucketFS) Writer(name URI) (io.WriteCloser, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	bucket, object := SplitGCPPath(name.Path)
	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	return wc, nil
}

func (fs *GCPBucketFS) Reader(name URI) (io.ReadCloser, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	bucket, object := SplitGCPPath(name.Path)
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func (fs *GCPBucketFS) Delete(name URI, recursive bool) error {
	return nil
}
func (fs *GCPBucketFS) Move(oldName, newName URI, recursive bool) error {
	return nil
}

func (fs *GCPBucketFS) Copy(oldName, newName URI, recursive bool) error {
	return nil
}

func (fs *GCPBucketFS) List(dir URI, recursive bool) ([]Node, error) {
	var files []Node
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	bucket, object := SplitGCPPath(dir.Path)
	var query storage.Query
	if !recursive {
		query = storage.Query{Prefix: object}
		query.Delimiter = "/"
		//Make sure object is a directory by adding a trailing slash
		if !strings.HasSuffix(object, "/") {
			query.Prefix = query.Prefix + "/"
		}
	}

	it := client.Bucket(bucket).Objects(ctx, &query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		if attrs.Name == object {
			continue
		}
		if attrs.Prefix != "" {
			// This is a 'folder'
			files = append(files, Node{Name: attrs.Prefix, IsDir: true})
		} else {
			// This is a file
			files = append(files, Node{Name: attrs.Name, IsDir: false})
		}

	}
	return files, nil
}

func (fs *GCPBucketFS) Get(path URI) (Node, error) {
	return Node{}, nil
}

func (fs *GCPBucketFS) Exists(path URI) (bool, error) {
	return false, nil
}

func (fs *GCPBucketFS) MkDir(path URI) (Node, error) {
	return Node{}, nil
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
