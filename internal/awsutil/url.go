package awsutil

import (
	"net/url"
	"strings"
)

// EscapePath escapes URL path according to AWS escaping rules.
//
// This func can be used to escape S3 object keys for HTTP access.
func EscapePath(path string) string {
	// AWS SDK v2 doesn't expose the rest.EscapePath function anymore
	// We need to implement the same behavior using url.PathEscape
	// AWS uses a custom escaping that doesn't escape forward slashes
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}
