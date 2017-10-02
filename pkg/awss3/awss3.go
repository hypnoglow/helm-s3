package awss3

import (
	"context"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

// FetchRaw downloads the file from URI and returns it in the form of byte slice.
// URI must be in the form of s3://bucket-name/key[/file.ext].
func FetchRaw(ctx context.Context, uri string, awsConfig *aws.Config) ([]byte, error) {
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new aws session")
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse uri %s", uri)
	}

	bucket, key := u.Host, strings.TrimPrefix(u.Path, "/")

	buf := &aws.WriteAtBuffer{}
	_, err = s3manager.NewDownloader(sess).DownloadWithContext(
		ctx,
		buf,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		return nil, errors.Wrap(err, "failed to download object from s3")
	}

	return buf.Bytes(), nil
}

// Upload uploads the file read from r to S3 by path uri. URI must be in the form
// of s3://bucket-name/key[/file.ext].
func Upload(ctx context.Context, uri string, r io.Reader, awsConfig *aws.Config) (string, error) {
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return "", errors.Wrap(err, "failed to create new aws session")
	}

	u, err := url.Parse(uri)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse uri %s", uri)
	}

	bucket, key := u.Host, strings.TrimPrefix(u.Path, "/")

	result, err := s3manager.NewUploader(sess).UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to upload file to s3")
	}

	return result.Location, nil
}
