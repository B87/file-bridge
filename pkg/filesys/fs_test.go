package filesys

import (
	"errors"
	"testing"
)

func TestParseURI(t *testing.T) {

	testCases := []struct {
		name    string
		in      string
		wantURI URI
		wantErr error
	}{
		{name: "local", in: "tmp.txt", wantURI: NewURI(LocalScheme, "tmp.txt"), wantErr: nil},
		{name: "gcp bucket", in: "gs://tmp.txt", wantURI: NewURI(GCPBucketScheme, "tmp.txt"), wantErr: nil},
		{name: "unknown", in: "moc://tmp.txt", wantURI: NewURI("moc", "tmp.txt"), wantErr: ErrUnknownScheme},
		{name: "invalid", in: "", wantURI: NewURI("", ""), wantErr: ErrInvalidURI},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotURI, gotErr := ParseURI(tc.in)
			if gotURI.Scheme != tc.wantURI.Scheme {
				t.Errorf("ParseURI(%v) = %v, want %v", tc.in, gotURI.Scheme, tc.wantURI.Scheme)
			}
			if gotURI.Path != tc.wantURI.Path {
				t.Errorf("ParseURI(%v) = %v, want %v", tc.in, gotURI.Path, tc.wantURI.Path)
			}
			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("ParseURI(%v) = %v, want %v", tc.in, gotErr, tc.wantErr)
			}
		})
	}

}
func TestCheckPrefix(t *testing.T) {

	testCases := []struct {
		name     string
		in       string
		wantFS   FS
		wantErr  error
		wantName string
	}{
		{name: "local", in: "tmp.txt", wantFS: NewLocalFS(), wantErr: nil, wantName: "tmp.txt"},
		{name: "gcp bucket", in: "gs://tmp.txt", wantFS: NewGCPBucketFS(), wantErr: nil, wantName: "tmp.txt"},
		{name: "unknown", in: "moc://tmp.txt", wantFS: nil, wantErr: ErrUnknownScheme, wantName: "tmp.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uri, err := ParseURI(tc.in)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("CheckPrefix(%v) = %v, want %v", tc.in, err, tc.wantErr)
			}
			if uri.Path != tc.wantName {
				t.Errorf("CheckPrefix(%v) = %v, want %v", tc.in, uri.Path, tc.wantName)
			}

		})
	}
}
