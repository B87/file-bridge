package manager

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/B87/file-bridge/pkg/filesystem"
	"github.com/B87/file-bridge/pkg/uri"
)

// Add tracer at package level
var tracer = otel.Tracer("github.com/B87/file-bridge/pkg/manager")

type Manager struct {
	defaultFSName string
	DefaultFS     filesystem.FS
	FsMap         map[string]filesystem.FS
}

// NewManager creates a new Manager and connects all filesystems
func NewManager(ctx context.Context, cfg *Config) (*Manager, error) {
	var defaultFS filesystem.FS
	var defaultFSName string
	for _, fsConfig := range cfg.FsConfigs {
		if fsConfig.Default {
			defaultFSName = fsConfig.Name
			defaultFS = cfg.FsMap[fsConfig.Name]
		}
	}

	return &Manager{FsMap: cfg.FsMap, DefaultFS: defaultFS, defaultFSName: defaultFSName}, nil
}

/*
Move moves a file from one filesystem to another.

Is implemented as a copy [src] [dst] followed by a delete [src].
*/
func (m *Manager) Move(ctx context.Context, src, dst uri.URI, recursive bool) error {
	ctx, span := tracer.Start(ctx, "Manager.Move",
		trace.WithAttributes(
			attribute.String("src", src.String()),
			attribute.String("dst", dst.String()),
			attribute.Bool("recursive", recursive),
		),
	)
	defer span.End()

	// Copy first
	if err := m.Copy(ctx, src, dst, recursive); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to copy during move operation: %w", err)
	}

	// Only delete if copy was successful
	if err := m.Delete(ctx, src); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete source after copy: %w", err)
	}

	return nil
}

/*
Lists the contents of a directory. If recursive is true, lists recursively.
*/
func (m *Manager) List(ctx context.Context, path uri.URI, recursive bool) (filesystem.NodeTree, error) {
	ctx, span := tracer.Start(ctx, "Manager.List",
		trace.WithAttributes(
			attribute.String("path", path.String()),
			attribute.Bool("recursive", recursive),
			attribute.String("scheme", path.Scheme),
		),
	)
	defer span.End()

	fs := m.schemeFS(path.Scheme)
	nodes, err := fs.List(ctx, path, recursive)
	if err != nil {
		span.RecordError(err)
		return filesystem.NodeTree{}, err
	}
	return nodes, nil
}

// Delete deletes a file or directory
func (m *Manager) Delete(ctx context.Context, path uri.URI) error {
	fs := m.schemeFS(path.Scheme)
	return fs.Delete(ctx, path)
}

// MkDir creates an empty directory
func (m *Manager) MkDir(ctx context.Context, path uri.URI) (filesystem.Node, error) {
	fs := m.schemeFS(path.Scheme)
	return fs.MkDir(ctx, path)
}

func (m *Manager) schemeFS(scheme string) filesystem.FS {
	if scheme == "" {
		return m.DefaultFS
	}
	fs, ok := m.FsMap[scheme]
	if !ok {
		panic(fmt.Sprintf("filesystem %s not found", scheme))
	}
	return fs
}
