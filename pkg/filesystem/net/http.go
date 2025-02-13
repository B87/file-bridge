package http

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/B87/file-bridge/pkg/filesystem"
	"github.com/B87/file-bridge/pkg/uri"
)

type HttpFS struct {
	Client http.Client
}

func NewHttpFS() *HttpFS {
	return &HttpFS{}
}

func (fs *HttpFS) Reader(ctx context.Context, uri uri.URI) (io.ReadCloser, error) {
	resp, err := fs.Client.Get(uri.String())
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (fs *HttpFS) Writer(ctx context.Context, uri uri.URI) (io.WriteCloser, error) {
	return nil, errors.New("cannot write over HTTP")
}

func (fs *HttpFS) List(ctx context.Context, uri uri.URI, recursive bool) (filesystem.NodeTree, error) {
	n, err := fs.Get(ctx, uri)
	if err != nil {
		return filesystem.NodeTree{}, err
	}
	return filesystem.NodeTree{Node: n}, nil
}

func (fs *HttpFS) Get(ctx context.Context, uri uri.URI) (filesystem.Node, error) {
	return filesystem.NewNode(uri, true), nil
}

func (fs *HttpFS) MkDir(ctx context.Context, uri uri.URI) (filesystem.Node, error) {
	return filesystem.Node{}, errors.New("cannot create a directory over HTTP")
}

func (fs *HttpFS) Delete(ctx context.Context, uri uri.URI) error {
	return errors.New("cannot delete over HTTP")
}

func (fs *HttpFS) Copy(ctx context.Context, src, dst uri.URI) error {
	return errors.New("cannot copy over HTTP")
}
