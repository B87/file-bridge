package filesystem

import "strings"

// SuffixSlash ensures that the given string ends with a slash. If not, it adds one.
func SuffixSlash(s string) string {
	if !strings.HasSuffix(s, "/") {
		s += "/"
	}
	return s
}

// SplitPath splits a path into a bucket and an object.
func SplitPath(path string) (bucket string, object string) {
	tree := strings.Split(path, "/")
	if len(tree) < 1 {
		return "", ""
	}
	bucket = tree[0]
	if len(tree) < 2 {
		return bucket, ""
	}
	object = strings.Join(tree[1:], "/")
	return bucket, object
}
