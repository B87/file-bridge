package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/schollz/progressbar/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/B87/file-bridge/pkg/filesystem"
	"github.com/B87/file-bridge/pkg/uri"
)

/*
Copy copies a file from one filesystem to another.

If both filesystems are the same, use the filesystem's copy method.

If the filesystems are different, perform a manual copy:
  - get the source file
  - create the destination file
*/
func (m *Manager) Copy(ctx context.Context, src, dst uri.URI, recursive bool) error {
	ctx, span := tracer.Start(ctx, "Manager.Copy",
		trace.WithAttributes(
			attribute.String("src", src.String()),
			attribute.String("dst", dst.String()),
			attribute.Bool("recursive", recursive),
		),
	)
	defer span.End()

	srcFS := m.schemeFS(src.Scheme)
	dstFS := m.schemeFS(dst.Scheme)
	// If the destination is a directory, add the source file name to the destination path
	srcNode, err := srcFS.Get(ctx, src)
	if err != nil && !errors.Is(err, filesystem.ErrNotFound) {
		return fmt.Errorf("failed to get source file %s: %w", src.String(), err)
	} else if errors.Is(err, filesystem.ErrNotFound) {
		return fmt.Errorf("source file %s does not exist", src.String())
	}

	// Create parent directories regardless of source type
	parentDir := path.Dir(dst.Path)
	if parentDir != "." && parentDir != "/" {
		mkdirURI := dst
		mkdirURI.Path = parentDir
		_, err := dstFS.MkDir(ctx, mkdirURI)
		if err != nil && !errors.Is(err, filesystem.ErrAlreadyExists) {
			return fmt.Errorf("failed to create parent directory %s: %w", mkdirURI.String(), err)
		}
	}

	if srcNode.IsDir {
		dst.Path = path.Join(dst.Path, srcNode.URI.Name)
		_, err := dstFS.MkDir(ctx, dst)
		if err != nil && !errors.Is(err, filesystem.ErrAlreadyExists) {
			return fmt.Errorf("failed to create destination directory %s: %w", dst.String(), err)
		}
	}
	// If the filesystems are the same, use the filesystem's copy method
	// we might get a better performance using the native copy method if exists
	if srcFS == dstFS {
		return srcFS.Copy(ctx, src, dst)
	} else {
		if srcNode.IsDir {
			// List the source directory and copy each file
			nodes, err := m.List(ctx, src, recursive)
			if err != nil {
				return err
			}

			for _, node := range nodes.Children {
				if !node.Node.IsDir {
					err := CopyFile(ctx, node.Node.URI, dst, srcFS, dstFS)
					if err != nil {
						return err
					}
				}
			}
		} else {
			// Just copy one file
			err := CopyFile(ctx, srcNode.URI, dst, srcFS, dstFS)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CopyFile(ctx context.Context, src, dst uri.URI, srcFS, dstFS filesystem.FS) error {
	ctx, span := tracer.Start(ctx, "Manager.CopyFile",
		trace.WithAttributes(
			attribute.String("src", src.String()),
			attribute.String("dst", dst.String()),
		),
	)
	defer span.End()

	// Modify the destination path to include the file name before opening files
	dst.Path = path.Join(dst.Path, src.Name)

	fmt.Println("Copying file", src.String(), "to", dst.String())

	srcFile, err := srcFS.Reader(ctx, src)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := dstFS.Writer(ctx, dst)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Get progress bar from context if available
	var bar *progressbar.ProgressBar
	if pb := ctx.Value("progressbar"); pb != nil {
		bar = pb.(*progressbar.ProgressBar)
	}

	// Use a buffer to copy in chunks and update progress
	buf := make([]byte, 32*1024) // 32KB buffer
	var copied int64
	for {
		n, err := srcFile.Read(buf)
		if err != nil && err != io.EOF {
			span.RecordError(err)
			return fmt.Errorf("failed to read from source: %w", err)
		}
		if n == 0 {
			break
		}

		if _, err := dstFile.Write(buf[:n]); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to write to destination: %w", err)
		}

		copied += int64(n)
		if bar != nil {
			bar.Add64(int64(n))
		}
	}

	span.SetAttributes(attribute.Int64("bytes_copied", copied))
	return nil
}
