package manager

import (
	"context"
	"errors"
	"fmt"

	"github.com/B87/file-bridge/pkg/filesystem"
	"github.com/B87/file-bridge/pkg/uri"
)

// SyncOneWay synchronizes the source filesystem with the destination filesystem
func (m *Manager) SyncOneWay(ctx context.Context, src, dst uri.URI, hashRepo URIHashStorage) error {
	// Get the source and destination filesystems
	srcFS := m.schemeFS(src.Scheme)
	dstFS := m.schemeFS(dst.Scheme)

	// Get the source and destination nodes
	srcNode, err := srcFS.Get(ctx, src)
	if err != nil {
		return err
	}
	dstNode, err := dstFS.Get(ctx, dst)
	if err != nil {
		return err
	}
	if !dstNode.IsDir {
		return fmt.Errorf("source is a directory but destination is a file")
	}

	if srcNode.IsDir {
		srcNodes, err := m.List(ctx, src, true)
		if err != nil {
			return err
		}

		// Sync files from source to destination
		for _, srcNode := range srcNodes.Children {
			if !srcNode.Node.IsDir {
				dstNode, err := dstFS.Get(ctx, srcNode.Node.URI)
				if errors.Is(err, filesystem.ErrNotFound) {
					// File does not exist in the destination
					err := CopyFile(ctx, srcNode.Node.URI, dst, srcFS, dstFS)
					if err != nil {
						return err
					}
				} else if err != nil {
					// Any other error getting the destination file
					return err
				} else {
					// File exists in the destination
					copyFileIfDifferent(ctx, srcNode.Node.URI, dstNode.URI, srcFS, dstFS, hashRepo)
				}
			}
		}
	} else {
		copyFileIfDifferent(ctx, src, dst, srcFS, dstFS, hashRepo)
	}

	return nil
}

func copyFileIfDifferent(ctx context.Context, src, dst uri.URI, srcFS, dstFS filesystem.FS, hashRepo URIHashStorage) error {
	// Get initial hashes
	srcHash, err := hashRepo.Get(src)
	if err != nil {
		return fmt.Errorf("failed to get source hash for %s: %w", src.String(), err)
	}

	dstHash, err := hashRepo.Get(dst)
	if err != nil {
		return fmt.Errorf("failed to get destination hash for %s: %w", dst.String(), err)
	}

	if srcHash != dstHash {
		// Copy file if hashes differ
		if err := CopyFile(ctx, src, dst, srcFS, dstFS); err != nil {
			return fmt.Errorf("failed to copy file from %s to %s: %w", src.String(), dst.String(), err)
		}

		// Generate and store new hashes
		srcHash, err := generateHash(ctx, src, srcFS)
		if err != nil {
			return fmt.Errorf("failed to generate source hash after copy for %s: %w", src.String(), err)
		}

		if err := hashRepo.Put(src, srcHash); err != nil {
			return fmt.Errorf("failed to store source hash for %s: %w", src.String(), err)
		}

		dstHash, err := generateHash(ctx, dst, dstFS)
		if err != nil {
			return fmt.Errorf("failed to generate destination hash after copy for %s: %w", dst.String(), err)
		}

		if err := hashRepo.Put(dst, dstHash); err != nil {
			return fmt.Errorf("failed to store destination hash for %s: %w", dst.String(), err)
		}
	}

	return nil
}

func generateHash(ctx context.Context, uri uri.URI, fs filesystem.FS) (string, error) {
	fsReader, err := fs.Reader(ctx, uri)
	if err != nil {
		return "", err
	}
	defer fsReader.Close()
	return "", nil
}
