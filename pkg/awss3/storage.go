package awss3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/provenance"

	"github.com/hypnoglow/helm-s3/pkg/awsutil"
)

var (
	ErrBucketNotFound = errors.New("bucket not found")
	ErrObjectNotFound = errors.New("object not found")
)

// New returns a new Storage.
func New() *Storage {
	return &Storage{}
}

// Storage provides an interface to work with AWS S3 objects by s3 protocol.
type Storage struct {
	session *session.Session
}

// Traverse traverses all charts in the repository.
// It writes an info item about every chart to items, and errors to errs.
// It always closes both channels when returns.
func (s *Storage) Traverse(ctx context.Context, repoURI string, items chan<- ChartInfo, errs chan<- error) {
	defer close(items)
	defer close(errs)

	if err := s.initSession(); err != nil {
		errs <- err
		return
	}

	bucket, key, err := parseURI(repoURI)
	if err != nil {
		errs <- err
		return
	}

	client := s3.New(s.session)

	var continuationToken *string
	for {
		listOut, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(key),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			errs <- errors.Wrap(err, "list s3 bucket objects")
			return
		}

		for _, obj := range listOut.Contents {
			if strings.Contains(*obj.Key, "/") {
				// This is a subfolder. Ignore it, because chart repository
				// is flat and cannot contain nested directories.
				continue
			}
			if *obj.Key == "index.yaml" {
				// Ignore the index itself.
				continue
			}

			metaOut, err := client.HeadObject(&s3.HeadObjectInput{
				Bucket: aws.String(bucket),
				Key:    obj.Key,
			})
			if err != nil {
				errs <- errors.Wrap(err, "head s3 object")
				return
			}

			reindexItem := ChartInfo{Filename: *obj.Key}

			// PROCESS THE OBJECT
			serializedChartMeta, hasMeta := metaOut.Metadata[strings.Title(metaChartMetadata)]
			chartDigest, hasDigest := metaOut.Metadata[strings.Title(metaChartDigest)]
			if !hasMeta || !hasDigest {
				// TODO: This is deprecated. Remove in next major release? Or not?
				// All charts pushed to the repository
				// since "reindex" command implementation should have these
				// meta fields.
				// But should we support the case when user manually uploads
				// the ch to the bucket? In this case, there will be no
				// such meta fields.

				// Anyway, in this case we have to download the ch file itself.
				objectOut, err := client.GetObject(&s3.GetObjectInput{
					Bucket: aws.String(bucket),
					Key:    obj.Key,
				})
				if err != nil {
					errs <- errors.Wrap(err, "get s3 object")
					return
				}

				buf := &bytes.Buffer{}
				tr := io.TeeReader(objectOut.Body, buf)

				ch, err := chartutil.LoadArchive(tr)
				objectOut.Body.Close()
				if err != nil {
					errs <- errors.Wrap(err, "load archive from s3 object")
					return
				}

				reindexItem.Meta = ch.Metadata
				reindexItem.Hash, err = provenance.Digest(buf)
				if err != nil {
					errs <- errors.WithMessage(err, "get chart hash")
					return
				}
			} else {
				reindexItem.Meta = &chart.Metadata{}
				if err := json.Unmarshal([]byte(*serializedChartMeta), reindexItem.Meta); err != nil {
					errs <- errors.Wrap(err, "unserialize chart meta")
					return
				}

				reindexItem.Hash = *chartDigest
			}

			// process meta and hash
			items <- reindexItem
		}

		// Decide if need to load more objects.
		if listOut.NextContinuationToken == nil {
			break
		}
		continuationToken = listOut.NextContinuationToken
	}
}

type ChartInfo struct {
	Meta     *chart.Metadata
	Filename string
	Hash     string
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
		if ae, ok := err.(awserr.Error); ok {
			if ae.Code() == s3.ErrCodeNoSuchBucket {
				return nil, ErrBucketNotFound
			}
			if ae.Code() == s3.ErrCodeNoSuchKey {
				return nil, ErrObjectNotFound
			}
		}
		return nil, errors.Wrap(err, "fetch object from s3")
	}

	return buf.Bytes(), nil
}

// Upload uploads the object read from r to S3 by path uri.
// uri must be in the form of s3 protocol: s3://bucket-name/key[...].
//
// Deprecated: use PutChart or PutIndex instead.
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

func (s *Storage) PutChart(ctx context.Context, uri string, r io.Reader, chartMeta, chartDigest string) (string, error) {
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
			Metadata: map[string]*string{
				metaChartMetadata: aws.String(chartMeta),
				metaChartDigest:   aws.String(chartDigest),
			},
		})
	if err != nil {
		return "", errors.Wrap(err, "upload object to s3")
	}

	return result.Location, nil
}

func (s *Storage) PutIndex(ctx context.Context, uri string, r io.Reader) error {
	if strings.HasPrefix(uri, "index.yaml") {
		return errors.New("uri must not contain \"index.yaml\" suffix, it appends automatically")
	}
	uri += "/index.yaml"

	if err := s.initSession(); err != nil {
		return err
	}

	bucket, key, err := parseURI(uri)
	if err != nil {
		return err
	}

	_, err = s3manager.NewUploader(s.session).UploadWithContext(
		ctx,
		&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   r,
		})
	if err != nil {
		return errors.Wrap(err, "upload index to S3 bucket")
	}

	return nil
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

const (
	metaChartMetadata = "chart-metadata"
	metaChartDigest   = "chart-digest"
)
