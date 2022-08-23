package awsutil

import "github.com/aws/aws-sdk-go/private/protocol/rest"

// EscapePath escapes URL path according to AWS escaping rules.
//
// This func can be used to escape S3 object keys for HTTP access.
func EscapePath(path string) string {
	return rest.EscapePath(path, true)
}
