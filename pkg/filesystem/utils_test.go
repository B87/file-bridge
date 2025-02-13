package filesystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuffixSlash(t *testing.T) {

	tests := []struct {
		name string
		want string
		arg  string
	}{
		{name: "empty", want: "/", arg: ""},
		{name: "slash", want: "/", arg: "/"},
		{name: "no slash dir", want: "/path/", arg: "/path"},
		{name: "slash dir", want: "/path/", arg: "/path/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SuffixSlash(tt.arg); got != tt.want {
				t.Errorf("SuffixSlash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitPath(t *testing.T) {

	tests := []struct {
		name           string
		path           string
		expectedBucket string
		expectedObject string
	}{
		{name: "Split path", path: "bucket/object", expectedBucket: "bucket", expectedObject: "object"},
		{name: "Split path with spaces", path: "bucket/object with spaces", expectedBucket: "bucket", expectedObject: "object with spaces"},
		{name: "Split multi folder level", path: "bucket/folder/object", expectedBucket: "bucket", expectedObject: "folder/object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, object := SplitPath(tt.path)
			assert.Equal(t, tt.expectedBucket, bucket)
			assert.Equal(t, tt.expectedObject, object)
		})
	}

}
