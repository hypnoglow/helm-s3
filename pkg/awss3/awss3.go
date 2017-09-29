package awss3

import (
	"context"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

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
