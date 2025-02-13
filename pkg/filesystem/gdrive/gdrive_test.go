package gdrive

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/B87/file-bridge/pkg/uri"
)

// Integration test example (requires actual Google Drive access)
func TestGDriveFS_Integration(t *testing.T) {
	driveID := os.Getenv("TEST_DRIVE_ID")
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	if driveID == "" || credentialsPath == "" {
		t.Skip("Skipping integration tests: TEST_DRIVE_ID or GOOGLE_APPLICATION_CREDENTIALS not set")
	}

	ctx := context.Background()
	svc, err := drive.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		t.Fatalf("Failed to create Drive service: %v", err)
	}

	fs := NewGDriveFS(svc, driveID)

	t.Run("file lifecycle", func(t *testing.T) {
		// Create a test file
		uri, _ := uri.Parse("/test/integration_test.txt")
		writer, err := fs.Writer(ctx, uri)
		assert.NoError(t, err)

		content := "integration test content"
		_, err = writer.Write([]byte(content))
		assert.NoError(t, err)
		err = writer.Close()
		assert.NoError(t, err)

		// Read the file back
		reader, err := fs.Reader(ctx, uri)
		assert.NoError(t, err)
		readContent, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, content, string(readContent))

		// Clean up
		err = fs.Delete(ctx, uri)
		assert.NoError(t, err)
	})
}
