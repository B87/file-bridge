package gdrive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"strings"
	"time"

	drive "google.golang.org/api/drive/v3"

	"github.com/B87/file-bridge/pkg/filesystem"
	"github.com/B87/file-bridge/pkg/uri"
)

const (
	gDriveFolderMimeType = "application/vnd.google-apps.folder"
)

// GDriveFS is a FileSystem implementation that uses a Shared Google Drive.
type GDriveFS struct {
	svc     *drive.Service
	driveID string
}

// GDriveFS implements the FileSystem interface.
var _ filesystem.FS = &GDriveFS{}

// NewGDriveFS creates a new GDriveFS.
func NewGDriveFS(svc *drive.Service, driveID string) *GDriveFS {
	return &GDriveFS{svc: svc, driveID: driveID}
}

// Writer returns an io.WriteCloser that, when closed, uploads the written data
// to the file indicated by name. If the file doesn't exist, it is created.
// For large files, you may want to handle chunked uploads or use the resumable upload API.
func (g *GDriveFS) Writer(ctx context.Context, name uri.URI) (io.WriteCloser, error) {

	// Ensure path is valid, find or create necessary folders, etc.
	fileName, folderID, err := g.ensurePathExists(ctx, name)
	if err != nil {
		return nil, err
	}

	// Check if a file with the same name already exists under folderID.
	fileID, err := g.findFileIDByName(ctx, fileName, folderID)
	var create bool
	if errors.Is(err, fs.ErrNotExist) {
		// File does not exist, create new
		create = true
	} else if err != nil {
		return nil, err
	}

	// We'll use a pipe so that we can stream data to Drive as it is written.
	pipeReader, pipeWriter := io.Pipe()

	// We’ll perform the upload in a separate goroutine.
	go func() {
		defer pipeReader.Close()

		if create {
			// Create a new file
			file := &drive.File{
				Name:    fileName,
				Parents: []string{folderID},
				DriveId: g.driveID, // needed if using a shared drive
			}
			_, err := g.svc.Files.Create(file).
				Media(pipeReader).
				SupportsAllDrives(true).
				Do()
			if err != nil {
				log.Printf("Writer: failed to create file: %v", err)
				_ = pipeWriter.CloseWithError(err)
			} else {
				_ = pipeWriter.Close()
			}
		} else {
			// Update existing file
			file := &drive.File{
				Name: fileName,
			}
			_, err := g.svc.Files.Update(fileID, file).
				Media(pipeReader).
				SupportsAllDrives(true).
				Do()
			if err != nil {
				log.Printf("Writer: failed to update file: %v", err)
				_ = pipeWriter.CloseWithError(err)
			} else {
				_ = pipeWriter.Close()
			}
		}
	}()

	// Return the pipeWriter so the caller can write to it.
	return pipeWriter, nil
}

func (g *GDriveFS) Reader(ctx context.Context, name uri.URI) (io.ReadCloser, error) {

	fileName, folderID, err := g.ensurePathExists(ctx, name, false /*createIfMissing*/)
	if err != nil {
		return nil, err
	}

	// Find the file ID by name in the folder
	fileID, err := g.findFileIDByName(ctx, fileName, folderID)
	if err != nil {
		return nil, err
	}

	// Download the file content
	res, err := g.svc.Files.Get(fileID).
		SupportsAllDrives(true).
		Download()
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return res.Body, nil
}

// Get fetches file/folder metadata from Google Drive and returns a Node struct.
func (g *GDriveFS) Get(ctx context.Context, name uri.URI) (filesystem.Node, error) {

	fileName, folderID, err := g.ensurePathExists(ctx, name, false /*createIfMissing*/)
	if err != nil {
		return filesystem.Node{}, err
	}

	fileID, err := g.findFileIDByName(ctx, fileName, folderID)
	if err != nil {
		return filesystem.Node{}, err
	}

	f, err := g.svc.Files.Get(fileID).
		SupportsAllDrives(true).
		Fields("id,name,mimeType,size,modifiedTime,parents").
		Do()
	if err != nil {
		return filesystem.Node{}, fmt.Errorf("failed to get file: %w", err)
	}
	t, err := time.Parse(time.RFC3339, f.ModifiedTime)
	if err != nil {
		return filesystem.Node{}, fmt.Errorf("failed to parse modified time: %w", err)
	}
	n := filesystem.Node{
		Name:    f.Name,
		Size:    int64(f.Size),
		Path:    name.String(),
		IsDir:   f.MimeType == gDriveFolderMimeType,
		ModTime: t,
	}

	return n, nil
}

// Delete removes a file or folder from Drive.
func (g *GDriveFS) Delete(ctx context.Context, name uri.URI) error {

	fileName, folderID, err := g.ensurePathExists(ctx, name, false /*createIfMissing*/)
	if err != nil {
		return err
	}
	fileID, err := g.findFileIDByName(ctx, fileName, folderID)
	if err != nil {
		return err
	}

	node, _ := g.Get(ctx, name)
	if node.IsDir {
		// recursively delete children
		nodes, err := g.List(ctx, name, false)
		if err != nil {
			return fmt.Errorf("failed to list folder before delete: %w", err)
		}
		for _, child := range nodes.Children {
			childURI, _ := uri.Parse(child.Node.Path + "/" + child.Node.Name)
			if err := g.Delete(ctx, childURI); err != nil {
				return err
			}
		}
	}

	// Delete the file/folder itself
	return g.svc.Files.Delete(fileID).
		SupportsAllDrives(true).
		Do()
}

// Copy duplicates a file in Drive. Overwrite controls if it should replace
// an existing file with the same name in the destination.
func (g *GDriveFS) Copy(ctx context.Context, src, dst uri.URI) error {

	// Identify source
	srcFileName, srcFolderID, err := g.ensurePathExists(ctx, src, false)
	if err != nil {
		return err
	}
	srcFileID, err := g.findFileIDByName(ctx, srcFileName, srcFolderID)
	if err != nil {
		return err
	}

	// Identify destination
	dstFileName, dstFolderID, err := g.ensurePathExists(ctx, dst)
	if err != nil {
		return err
	}

	// Overwrite, delete if the dest file exists
	if existingDestFileID, err := g.findFileIDByName(ctx, dstFileName, dstFolderID); err == nil {
		_ = g.svc.Files.Delete(existingDestFileID).
			SupportsAllDrives(true).
			Do()
	}

	// Perform the copy
	newFile := &drive.File{
		Name:    dstFileName,
		Parents: []string{dstFolderID},
		DriveId: g.driveID,
	}
	_, err = g.svc.Files.Copy(srcFileID, newFile).
		SupportsAllDrives(true).
		Do()
	return err
}

// List returns items in the folder indicated by name. If recursive is true,
// the listing should include all descendants (this example does a shallow listing).
func (g *GDriveFS) List(ctx context.Context, name uri.URI, recursive bool) (filesystem.NodeTree, error) {
	var tree filesystem.NodeTree
	_, folderID, err := g.ensurePathExists(ctx, name, false /*createIfMissing*/)
	if err != nil {
		return tree, err
	}

	q := fmt.Sprintf("trashed=false and '%s' in parents", folderID)
	fileList, err := g.svc.Files.List().
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Corpora("drive").
		DriveId(g.driveID).
		Q(q).
		Fields("files(id, name, mimeType, size, modifiedTime)").
		Do()
	if err != nil {
		return tree, fmt.Errorf("failed to list drive folder: %w", err)
	}

	for _, f := range fileList.Files {
		n := filesystem.Node{
			Name:  f.Name,
			IsDir: f.MimeType == gDriveFolderMimeType,
		}
		if f.Size > 0 {
			n.Size = int64(f.Size)
		}
		if f.ModifiedTime != "" {
			t, _ := time.Parse(time.RFC3339, f.ModifiedTime)
			n.ModTime = t
		}
		// For constructing node path, you'd typically join folder path + f.Name
		// But we don't have the folder's path handy here - you can maintain it yourself or adapt logic.
		n.Path = name.String()

		tree.Children = append(tree.Children, filesystem.NodeTree{Node: n})

		// If recursive and this is a folder, you’d recursively call List on that folder.
		// That is more complex and might need care to avoid excessive API calls.
	}

	return filesystem.NodeTree{
		Node:     filesystem.Node{Name: "", Path: name.String(), IsDir: true},
		Children: tree.Children,
	}, nil
}

// MkDir creates the specified directory in the Drive. If intermediate directories
// do not exist, they are created. Returns a Node representing the created folder.
func (g *GDriveFS) MkDir(ctx context.Context, name uri.URI) (filesystem.Node, error) {

	fileName, parentID, err := g.ensurePathExists(ctx, name, true /*createIfMissing*/, true /*folderOnly*/)
	if err != nil {
		return filesystem.Node{}, err
	}

	// See if it already exists as a folder
	fileID, err := g.findFileIDByName(ctx, fileName, parentID)
	if err == nil {
		// It exists. Return its Node
		f, err := g.svc.Files.Get(fileID).
			SupportsAllDrives(true).
			Fields("id,name,modifiedTime").
			Do()
		if err != nil {
			return filesystem.Node{}, err
		}
		t, _ := time.Parse(time.RFC3339, f.ModifiedTime)
		return filesystem.Node{
			Name:    f.Name,
			Path:    name.String(),
			IsDir:   true,
			ModTime: t,
		}, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		// Some other error
		return filesystem.Node{}, err
	}

	// Create the folder
	file := &drive.File{
		Name:     fileName,
		MimeType: gDriveFolderMimeType,
		Parents:  []string{parentID},
		DriveId:  g.driveID,
	}
	created, err := g.svc.Files.Create(file).
		SupportsAllDrives(true).
		Fields("id,name,modifiedTime").
		Do()
	if err != nil {
		return filesystem.Node{}, fmt.Errorf("failed to create folder: %w", err)
	}
	t, _ := time.Parse(time.RFC3339, created.ModifiedTime)
	return filesystem.Node{
		Name:    created.Name,
		Path:    name.String(),
		IsDir:   true,
		ModTime: t,
	}, nil
}

//
// Helper functions
//

// ensurePathExists breaks down the URI path into folders, ensuring each one exists.
// Returns the final file/folder name (leaf) plus the ID of its parent folder.
// If createIfMissing is false, it returns an error if any intermediate folder does not exist.
// If folderOnly is true, treat the final segment as a folder too (e.g. for MkDir).
func (g *GDriveFS) ensurePathExists(ctx context.Context, name uri.URI, opts ...interface{}) (fileName, parentID string, err error) {
	// Parse optional args
	createIfMissing := true
	folderOnly := false
	if len(opts) > 0 {
		if v, ok := opts[0].(bool); ok {
			createIfMissing = v
		}
	}
	if len(opts) > 1 {
		if v, ok := opts[1].(bool); ok {
			folderOnly = v
		}
	}

	segments := strings.Split(strings.Trim(name.Path, "/"), "/")
	if len(segments) == 0 {
		return "", "", fmt.Errorf("invalid path: %q", name.Path)
	}
	// If folderOnly == true, the final segment is a folder, so we handle everything else as folders as well.
	// If folderOnly == false, the last segment is the file name.
	endIdx := len(segments) - 1
	if folderOnly {
		endIdx = len(segments)
	}

	// Start from "root" for the given drive. For a shared drive, the root is the drive's top-level.
	parentID, err = g.getRootFolderID(ctx)
	if err != nil {
		return "", "", err
	}

	// Traverse intermediate folders
	for i := 0; i < endIdx; i++ {
		subFolderName := segments[i]
		// See if subFolderName exists
		fID, findErr := g.findFileIDByName(ctx, subFolderName, parentID)
		if findErr != nil {
			if !errors.Is(findErr, fs.ErrNotExist) {
				return "", "", findErr
			}
			// If createIfMissing is false, we can't proceed
			if !createIfMissing {
				return "", "", fs.ErrNotExist
			}
			// Create it
			newFolder := &drive.File{
				Name:     subFolderName,
				MimeType: gDriveFolderMimeType,
				Parents:  []string{parentID},
				DriveId:  g.driveID,
			}
			created, err := g.svc.Files.Create(newFolder).
				SupportsAllDrives(true).
				Fields("id").
				Do()
			if err != nil {
				return "", "", err
			}
			fID = created.Id
		}
		parentID = fID
	}

	// Return the final name (which might be a file or folder name) plus the ID of its parent folder
	if folderOnly {
		// There's no final "file name" because everything was a folder
		return segments[len(segments)-1], parentID, nil
	} else {
		return segments[len(segments)-1], parentID, nil
	}
}

// findFileIDByName looks for a file with the given name under parentFolderID.
// Returns fs.ErrNotExist if none is found.
func (g *GDriveFS) findFileIDByName(_ context.Context, name, parentFolderID string) (string, error) {
	q := fmt.Sprintf("trashed=false and name='%s' and '%s' in parents", escapeQuote(name), parentFolderID)
	res, err := g.svc.Files.List().
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Corpora("drive").
		DriveId(g.driveID).
		Q(q).
		Fields("files(id, name)").
		Do()
	if err != nil {
		return "", err
	}
	if len(res.Files) == 0 {
		return "", fs.ErrNotExist
	}
	// For simplicity, assume the first match is correct
	return res.Files[0].Id, nil
}

// getRootFolderID returns the top-level folder ID to use. For My Drive, this is "root".
// For a shared (team) drive, you can either hard-code the drive's root folder ID, or simply
// use the drive ID in your queries. If the Drive ID is known, you can call the API to fetch
// the root folder ID (the "drive" resource).
func (g *GDriveFS) getRootFolderID(_ context.Context) (string, error) {
	if g.driveID == "" {
		// My Drive: the special folder ID "root" references the user's root.
		return "root", nil
	}
	// For a shared drive, let's fetch the drive's root
	dr, err := g.svc.Drives.Get(g.driveID).Do()
	if err != nil {
		return "", fmt.Errorf("failed to get shared drive info: %w", err)
	}
	return dr.Id, nil // in many cases, dr.Id can be used as the root. Some older references used dr.RootFolderId
}

// escapeQuote wraps a string in quotes properly for Drive queries
func escapeQuote(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}
