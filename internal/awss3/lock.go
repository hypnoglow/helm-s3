package awss3

import "fmt"

// LockID returns the lockID for the bucket specified by the given URI.
func LockID(uri string) string {
	bucket, _, err := parseURI(uri)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("lock/%s", bucket)
}
