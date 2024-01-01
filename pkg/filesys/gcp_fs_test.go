package filesys

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitGCPPath(t *testing.T) {

	tests := []struct {
		name           string
		path           string
		expectedBucket string
		expectedObject string
	}{
		{name: "Split GCP path", path: "bucket/object", expectedBucket: "bucket", expectedObject: "object"},
		{name: "Split GCP path with spaces", path: "bucket/object with spaces", expectedBucket: "bucket", expectedObject: "object with spaces"},
		{name: "Split GCP multi folder level", path: "bucket/folder/object", expectedBucket: "bucket", expectedObject: "folder/object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, object := splitGCPPath(tt.path)
			assert.Equal(t, tt.expectedBucket, bucket)
			assert.Equal(t, tt.expectedObject, object)
		})
	}

}
