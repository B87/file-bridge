package uri

import (
	"errors"
	"path"
	"regexp"
)

var (
	ErrInvalidURI    = errors.New("invalid URI")
	ErrUnknownScheme = errors.New("unknown scheme")
)

const (
	EmptyScheme     string = ""
	LocalScheme     string = "local"
	GCPBucketScheme string = "gs"
	AWSScheme       string = "s3"
	DriveScheme     string = "gdrive"
	HTTPScheme      string = "http"
	HTTPSScheme     string = "https"
)

// ValidScheme is a stub that you can fill with your own logic.
// For example, you might allow only "local", "s3", "gs", "https", etc.
func ValidScheme(s string) bool {
	return s == EmptyScheme || s == LocalScheme || s == GCPBucketScheme ||
		s == AWSScheme || s == DriveScheme || s == HTTPScheme || s == HTTPSScheme
}

// URI represents a resource identifier, including optional scheme,
// path, query string, fragment, and a 'Name' derived from the path base.
type URI struct {
	// Scheme of the resource: "s3", "gs", "db", "local", etc.
	Scheme string

	// Path portion of the resource if no scheme is given,
	// this may include a leading slash for local paths (e.g. "/tmp/foo").
	// It should always be a full path, not a relative path.
	Path string

	// Query is the optional query string (everything after '?' but before '#').
	Query string

	// Fragment is the optional fragment (everything after '#').
	Fragment string

	// Name is typically path.Base(Path). For file-like resources,
	// this is the "file name" portion.
	Name string
}

// String reconstructs a URI-like string including the scheme:// if present.
func (u URI) String() string {
	out := ""
	if u.Scheme != "" {
		out = u.Scheme + "://"
	}
	out += u.Path
	if u.Query != "" {
		out += "?" + u.Query
	}
	if u.Fragment != "" {
		out += "#" + u.Fragment
	}
	return out
}

// Explanation of groups:
// ^                         start of string
// (?:(\w+):\/\/)?           optional group capturing the scheme (alphanumeric+_ only), followed by ://
// ([^?#]+)                  the "path" is any characters until ? or # or end
// (?:\?([^#]+))?            optional group capturing everything after '?' until '#' or end
// (?:#(.+))?                optional group capturing everything after '#'
// $                         end of string
var parseRE = regexp.MustCompile(`^(?:(\w+):\/\/)?([^?#]+)(?:\?([^#]+))?(?:#(.+))?$`)

func Parse(raw string) (URI, error) {
	matches := parseRE.FindStringSubmatch(raw)
	if matches == nil || len(matches) < 5 {
		// Matches: [full, scheme, path, query, fragment]
		return URI{Path: raw}, ErrInvalidURI
	}

	scheme := matches[1] // might be empty
	thePath := matches[2]
	query := matches[3]    // might be empty
	fragment := matches[4] // might be empty

	// Validate the scheme if present.
	if !ValidScheme(scheme) {
		return URI{Scheme: scheme, Path: thePath, Query: query, Fragment: fragment}, ErrUnknownScheme
	}

	// If no scheme is given, assume local.
	if scheme == EmptyScheme {
		scheme = LocalScheme
	}

	// For the name, we typically use path.Base of the path portion.
	name := path.Base(thePath)
	// If path ends in '/', path.Base will return "".
	// That might be okay for directories, or you can handle that logic differently.

	return URI{
		Scheme:   scheme,
		Path:     thePath,
		Query:    query,
		Fragment: fragment,
		Name:     name,
	}, nil
}

func MustParse(raw string) URI {
	uri, err := Parse(raw)
	if err != nil {
		panic(err)
	}
	return uri
}
