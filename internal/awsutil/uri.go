package awsutil

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// ParseURI returns bucket, key and region from URIs like:
// - s3://bucket-name/dir
// - s3://bucket-name/dir/file.ext
// - s3://region@bucket-name/dir/file.ext
func ParseURI(uri string) (bucket, key, region string, err error) {
	if !strings.HasPrefix(uri, "s3://") {
		return "", "", "", fmt.Errorf("uri %s protocol is not s3", uri)
	}

	u, err := url.Parse(uri)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "parse uri %s", uri)
	}

	bucket, key, region = u.Host, strings.TrimPrefix(u.Path, "/"), u.User.Username()
	return bucket, key, region, nil
}
