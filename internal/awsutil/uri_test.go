package awsutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNonRegionalURI(t *testing.T) {
	// ParseURI returns bucket, key and region from URIs like:
	// - s3://bucket-name/dir
	// - s3://bucket-name/dir/file.ext
	// - s3://region@bucket-name/dir/file.ext

	uri := "s3://bucket-name/dir"
	wantBucket, wantKey, wantRegion := "bucket-name", "dir", ""

	bucket, key, region, err := ParseURI(uri)
	assert.NoError(t, err)

	assert.Equal(t, wantBucket, bucket)
	assert.Equal(t, wantKey, key)
	assert.Equal(t, wantRegion, region)
}

func TestParseRegionalURI(t *testing.T) {
	// ParseURI returns bucket, key and region from URIs like:
	// - s3://bucket-name/dir
	// - s3://bucket-name/dir/file.ext
	// - s3://region@bucket-name/dir/file.ext
	uri := "s3://us-west-2@bucket-name/dir"
	wantBucket, wantKey, wantRegion := "bucket-name", "dir", "us-west-2"

	bucket, key, region, err := ParseURI(uri)
	assert.NoError(t, err)

	assert.Equal(t, wantBucket, bucket)
	assert.Equal(t, wantKey, key)
	assert.Equal(t, wantRegion, region)
}
