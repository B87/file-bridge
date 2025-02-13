package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/B87/file-bridge/pkg/filesystem"
	"github.com/B87/file-bridge/pkg/uri"
)

type AWSFS struct {
	session *session.Session
	svc     *s3.S3
}

func NewAWSFS() *AWSFS { return &AWSFS{} }

func (fs *AWSFS) Connect(ctx context.Context) error {
	region := os.Getenv("AWS_DEFAULT_REGION")
	endpoint := os.Getenv("AWS_ENDPOINT")
	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(endpoint),
		Region:   aws.String(region),
	})
	if err != nil {
		return err
	}
	fs.session = sess
	fs.svc = s3.New(sess)
	return nil
}
func (AWSFS) Disconnect(ctx context.Context) error { return nil }

// Writer returns a WriteCloser for the given URI.
func (fs *AWSFS) Writer(ctx context.Context, name uri.URI) (io.WriteCloser, error) {
	bucket, object := filesystem.SplitPath(name.Path)
	writer := NewS3WriteCloser(fs.session, bucket, object)
	return writer, nil

}

// Reader returns a ReadCloser for the given URI.
func (fs *AWSFS) Reader(ctx context.Context, name uri.URI) (io.ReadCloser, error) {
	bucket, object := filesystem.SplitPath(name.Path)
	resp, err := fs.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s: %v", filesystem.ErrFileList, object, err)
	}
	return resp.Body, nil
}

// Delete deletes the s3 file or directory at the given URI.
func (fs *AWSFS) Delete(ctx context.Context, name uri.URI) error {
	bucket, object := filesystem.SplitPath(name.Path)
	_, err := fs.svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	if err != nil {
		return fmt.Errorf("%w %s: %v", filesystem.ErrFileDelete, object, err)
	}
	return nil
}

// Copy copies the file or directory at the old URI to the new URI.
func (fs *AWSFS) Copy(ctx context.Context, oldName, newName uri.URI) error {
	oldB, oldO := filesystem.SplitPath(oldName.Path)
	newB, newO := filesystem.SplitPath(newName.Path)
	_, err := fs.svc.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(newB),
		CopySource: aws.String(oldB + "/" + oldO),
		Key:        aws.String(newO),
	})
	if err != nil {
		return fmt.Errorf("%w %s: %v", filesystem.ErrFileCopy, oldName, err)
	}
	return nil
}

// List returns a list of Nodes for the given URI.
func (fs *AWSFS) List(ctx context.Context, dir uri.URI, recursive bool) ([]filesystem.Node, error) {
	var nodes []filesystem.Node
	bucket, object := filesystem.SplitPath(dir.Path)
	var delimiter *string
	if recursive {
		delimiter = nil
	} else {
		delimiter = aws.String("/")
	}
	resp, err := fs.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Delimiter: delimiter,
		Prefix:    aws.String(filesystem.SuffixSlash(object)),
	})
	for _, item := range resp.Contents {
		nodes = append(nodes, filesystem.NewNode(buildS3URI(*item.Key), isDir(item)))
	}
	if err != nil {
		return nodes, fmt.Errorf("%w %s: %v", filesystem.ErrFileList, dir.Path, err)
	}
	return nodes, nil
}

// Get returns a Node for the given URI.
func (fs *AWSFS) Get(ctx context.Context, path uri.URI) (filesystem.Node, error) {
	bucket, object := filesystem.SplitPath(path.Path)
	resp, err := fs.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(filesystem.SuffixSlash(object)),
	})
	if err != nil {
		return filesystem.Node{URI: path}, err
	}
	for _, item := range resp.Contents {
		if *item.Key == object {
			return filesystem.NewNode(buildS3URI(*item.Key), isDir(item)), nil

		}
	}
	return filesystem.Node{}, nil
}

// MkDir creates a directory at the given URI.
func (fs *AWSFS) MkDir(ctx context.Context, path uri.URI) (filesystem.Node, error) {
	bucket, object := filesystem.SplitPath(path.Path)
	err := fs.PutObject(bucket, object, []byte(""))
	return filesystem.NewNode(path, true), err
}

func (fs *AWSFS) PutObject(bucket, objectKey string, content []byte) error {
	_, err := fs.svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(content),
	})
	return err
}

type S3WriteCloser struct {
	buffer  bytes.Buffer
	session *session.Session
	bucket  string
	key     string
	closed  bool
}

func NewS3WriteCloser(sess *session.Session, bucket, key string) *S3WriteCloser {
	return &S3WriteCloser{
		session: sess,
		bucket:  bucket,
		key:     key,
	}
}

func (w *S3WriteCloser) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, errors.New("write to a closed writer")
	}
	return w.buffer.Write(p)
}

func (w *S3WriteCloser) Close() error {
	if w.closed {
		return errors.New("close a closed writer")
	}
	w.closed = true
	uploader := s3manager.NewUploader(w.session)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(w.bucket),
		Key:    aws.String(w.key),
		Body:   bytes.NewReader(w.buffer.Bytes()),
	})
	return err
}

func buildS3URI(object string) uri.URI {
	return uri.URI{Scheme: uri.AWSScheme, Path: object}
}

func isDir(o *s3.Object) bool {
	return *o.Size == 0 && strings.HasSuffix(*o.Key, "/")
}
