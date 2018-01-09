package awss3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/pkg/awsutil"
)

// New returns a new Storage.
func New() *Storage {
	return &Storage{}
}

// Storage provides an interface to work with AWS S3 objects by s3 protocol.
type Storage struct {
	session *session.Session
}

// FetchRaw downloads the object from URI and returns it in the form of byte slice.
// uri must be in the form of s3 protocol: s3://bucket-name/key[...].
func (s *Storage) FetchRaw(ctx context.Context, uri string) ([]byte, error) {
	if err := s.initSession(); err != nil {
		return nil, err
	}

	bucket, key, err := parseURI(uri)
	if err != nil {
		return nil, err
	}

	buf := &aws.WriteAtBuffer{}
	_, err = s3manager.NewDownloader(s.session).DownloadWithContext(
		ctx,
		buf,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		return nil, errors.Wrap(err, "fetch object from s3")
	}

	return buf.Bytes(), nil
}

// Upload uploads the object read from r to S3 by path uri.
// uri must be in the form of s3 protocol: s3://bucket-name/key[...].
func (s *Storage) Upload(ctx context.Context, uri string, r io.Reader) (string, error) {
	if err := s.initSession(); err != nil {
		return "", err
	}

	bucket, key, err := parseURI(uri)
	if err != nil {
		return "", err
	}

	result, err := s3manager.NewUploader(s.session).UploadWithContext(
		ctx,
		&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   r,
		})
	if err != nil {
		return "", errors.Wrap(err, "upload object to s3")
	}

	return result.Location, nil
}

// Delete deletes the object by uri.
// uri must be in the form of s3 protocol: s3://bucket-name/key[...].
func (s *Storage) Delete(ctx context.Context, uri string) error {
	if err := s.initSession(); err != nil {
		return err
	}

	bucket, key, err := parseURI(uri)
	if err != nil {
		return err
	}

	_, err = s3.New(s.session).DeleteObjectWithContext(
		ctx,
		&s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		return errors.Wrap(err, "delete object from s3")
	}

	return nil
}

func (s *Storage) initSession() (err error) {
	if s.session != nil {
		return nil
	}

	s.session, err = awsutil.Session()
	return errors.Wrap(err, "init aws session")
}

// parseURI returns bucket and key from URIs like:
// - s3://bucket-name/dir
// - s3://bucket-name/dir/file.ext
func parseURI(uri string) (bucket, key string, err error) {
	if !strings.HasPrefix(uri, "s3://") {
		return "", "", fmt.Errorf("uri %s protocol is not s3", uri)
	}

	u, err := url.Parse(uri)
	if err != nil {
		return "", "", errors.Wrapf(err, "parse uri %s", uri)
	}

	bucket, key = u.Host, strings.TrimPrefix(u.Path, "/")
	return bucket, key, nil
}
