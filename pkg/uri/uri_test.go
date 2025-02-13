package uri

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
)

func TestParseURI(t *testing.T) {
	testCases := []struct {
		name     string
		raw      string
		expected URI
		wantErr  error
	}{
		{
			name:     "empty scheme (local path)",
			raw:      "/home/user/file.txt",
			expected: URI{Scheme: LocalScheme, Path: "/home/user/file.txt", Name: "file.txt"},
		},
		{
			name:     "local scheme",
			raw:      "file://home/user/file.txt",
			expected: URI{Scheme: LocalScheme, Path: "home/user/file.txt", Name: "file.txt"},
		},
		{
			name:     "s3 scheme",
			raw:      "s3://my-bucket/path/to/object.json",
			expected: URI{Scheme: AWSScheme, Path: "my-bucket/path/to/object.json", Name: "object.json"},
		},
		{
			name:     "gs scheme",
			raw:      "gs://my-bucket/path/to/object.txt",
			expected: URI{Scheme: GCPBucketScheme, Path: "my-bucket/path/to/object.txt", Name: "object.txt"},
		},
		{
			name:     "https scheme",
			raw:      "https://example.com/path/to/resource.pdf",
			expected: URI{Scheme: HTTPSScheme, Path: "example.com/path/to/resource.pdf", Name: "resource.pdf"},
		},
		{
			name:     "http scheme",
			raw:      "http://example.com/path/to/resource.html",
			expected: URI{Scheme: HTTPScheme, Path: "example.com/path/to/resource.html", Name: "resource.html"},
		},
		{
			name:     "gdrive scheme",
			raw:      "gdrive://folder-id/document.doc",
			expected: URI{Scheme: DriveScheme, Path: "folder-id/document.doc", Name: "document.doc"},
		},
		{
			name:     "with query parameters",
			raw:      "https://example.com/path?key=value&foo=bar",
			expected: URI{Scheme: HTTPSScheme, Path: "example.com/path", Query: "key=value&foo=bar", Name: "path"},
		},
		{
			name:     "with fragment",
			raw:      "https://example.com/doc.html#section1",
			expected: URI{Scheme: HTTPSScheme, Path: "example.com/doc.html", Fragment: "section1", Name: "doc.html"},
		},
		{
			name:    "invalid scheme",
			raw:     "invalid://example.com/path",
			wantErr: ErrUnknownScheme,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uri, err := Parse(tc.raw)
			if tc.wantErr != nil {
				if err != tc.wantErr {
					t.Errorf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if uri != tc.expected {
				t.Errorf("expected %+v, got %+v", tc.expected, uri)
			}
		})
	}
}

func TestValidScheme(t *testing.T) {
	testCases := []struct {
		name     string
		scheme   string
		expected bool
	}{
		{"empty scheme", "", true},
		{"s3 scheme", "s3", true},
		{"gs scheme", "gs", true},
		{"https scheme", "https", true},
		{"http scheme", "http", true},
		{"gdrive scheme", "gdrive", true},
		{"invalid scheme", "ftp", false},
		{"invalid scheme 2", "ssh", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidScheme(tc.scheme)
			if result != tc.expected {
				t.Errorf("ValidScheme(%q) = %v, want %v", tc.scheme, result, tc.expected)
			}
		})
	}
}

func TestURIString(t *testing.T) {
	testCases := []struct {
		name     string
		uri      URI
		expected string
	}{
		{"empty scheme", URI{Path: "/home/user/file.txt"}, "/home/user/file.txt"},
		{"with scheme", URI{Scheme: "s3", Path: "my-bucket/path/to/object.json"}, "s3://my-bucket/path/to/object.json"},
		{"with query", URI{Scheme: "https", Path: "example.com/path", Query: "key=value&foo=bar"}, "https://example.com/path?key=value&foo=bar"},
		{"with fragment", URI{Scheme: "https", Path: "example.com/path", Fragment: "section1"}, "https://example.com/path#section1"},
		{"gdrive", URI{Scheme: "gdrive", Path: "folder-id/document.doc"}, "gdrive://folder-id/document.doc"},
		{"gs", URI{Scheme: "gs", Path: "my-bucket/path/to/object.txt"}, "gs://my-bucket/path/to/object.txt"},
		{"s3", URI{Scheme: "s3", Path: "my-bucket/path/to/object.json"}, "s3://my-bucket/path/to/object.json"},
		{"http", URI{Scheme: "http", Path: "example.com/path/to/resource.html"}, "http://example.com/path/to/resource.html"},
		{"https", URI{Scheme: "https", Path: "example.com/path/to/resource.pdf"}, "https://example.com/path/to/resource.pdf"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.uri.String()
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestURIJSON(t *testing.T) {
	tests := []struct {
		name    string
		uri     URI
		want    string
		wantErr bool
	}{
		{
			name: "local file",
			uri:  URI{Scheme: "file", Path: "/path/to/file.txt"},
			want: `"file:///path/to/file.txt"`,
		},
		{
			name: "s3 with query and fragment",
			uri:  URI{Scheme: "s3", Path: "bucket/object", Query: "version=1", Fragment: "part1"},
			want: `"s3://bucket/object?version=1#part1"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			got, err := json.Marshal(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %v, want %v", string(got), tt.want)
			}

			// Test unmarshaling
			var u URI
			err = json.Unmarshal([]byte(tt.want), &u)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if u.String() != tt.uri.String() {
				t.Errorf("UnmarshalJSON() = %v, want %v", u, tt.uri)
			}
		})
	}
}

func TestURISQL(t *testing.T) {
	tests := []struct {
		name    string
		uri     URI
		value   interface{} // for Scan testing
		want    driver.Value
		wantErr bool
	}{
		{
			name:  "local file",
			uri:   URI{Scheme: "file", Path: "/path/to/file.txt"},
			value: "file:///path/to/file.txt",
			want:  "file:///path/to/file.txt",
		},
		{
			name:  "s3 path",
			uri:   URI{Scheme: "s3", Path: "bucket/object"},
			value: []byte("s3://bucket/object"),
			want:  "s3://bucket/object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Value()
			got, err := tt.uri.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}

			// Test Scan()
			var u URI
			err = u.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if u.String() != tt.uri.String() {
				t.Errorf("Scan() = %v, want %v", u, tt.uri)
			}
		})
	}
}
