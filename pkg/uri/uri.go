package uri

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"path"
	"regexp"
)

var (
	ErrInvalidURI    = errors.New("invalid URI")
	ErrUnknownScheme = errors.New("unknown scheme")
)

const (
	EmptyScheme     string = ""
	LocalScheme     string = "file"
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

// MarshalJSON implements json.Marshaler
func (u URI) MarshalJSON() ([]byte, error) {
	return []byte(`"` + u.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (u *URI) UnmarshalJSON(data []byte) error {
	// Remove quotes
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errors.New("invalid JSON string for URI")
	}
	str := string(data[1 : len(data)-1])

	parsed, err := Parse(str)
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}

// MarshalText implements encoding.TextMarshaler
func (u URI) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (u *URI) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}

// Value implements the driver.Valuer interface
func (u URI) Value() (driver.Value, error) {
	return u.String(), nil
}

// Scan implements the sql.Scanner interface
func (u *URI) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		parsed, err := Parse(v)
		if err != nil {
			return err
		}
		*u = parsed
		return nil
	case []byte:
		parsed, err := Parse(string(v))
		if err != nil {
			return err
		}
		*u = parsed
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into URI", value)
	}
}
