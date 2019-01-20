package awss3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
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
)

const (
	// selects serverside encryption for bucket
	awsS3encryption = "AWS_S3_SSE"
)

var (
	// ErrBucketNotFound signals that a bucket was not found.
	ErrBucketNotFound = errors.New("bucket not found")

	// ErrObjectNotFound signals that an object was not found.
	ErrObjectNotFound = errors.New("object not found")
)

// New returns a new Storage.
func New(session *session.Session) *Storage {
	return &Storage{session: session}
}

// Returns desired encryption
func getSSE() *string {
	sse := os.Getenv(awsS3encryption)
	if sse == "" {
		return nil
	}
	return &sse
}

// Storage provides an interface to work with AWS S3 objects by s3 protocol.
type Storage struct {
	session *session.Session
}

// Traverse traverses all charts in the repository.
func (s *Storage) Traverse(ctx context.Context, repoURI string) (<-chan ChartInfo, <-chan error) {
	charts := make(chan ChartInfo, 1)
	errs := make(chan error, 1)
	go s.traverse(ctx, repoURI, charts, errs)
	return charts, errs
}

// traverse traverses all charts in the repository.
// It writes an info item about every chart to items, and errors to errs.
// It always closes both channels when returns.
func (s *Storage) traverse(ctx context.Context, repoURI string, items chan<- ChartInfo, errs chan<- error) {
	defer close(items)
	defer close(errs)

	bucket, prefixKey, err := parseURI(repoURI)
	if err != nil {
		errs <- err
		return
	}

	client := s3.New(s.session)

	var continuationToken *string
	for {
		listOut, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefixKey),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			errs <- errors.Wrap(err, "list s3 bucket objects")
			return
		}

		for _, obj := range listOut.Contents {
			// We need to make object key relative to repo root.
			key := strings.TrimPrefix(*obj.Key, prefixKey)
			// Additionally trim prefix slash if exists, because repos can be:
			// s3://bucket/repo/subdir OR s3://bucket/repo/subdir/
			key = strings.TrimPrefix(key, "/")

			if strings.Contains(key, "/") {
				// This is a subfolder. Ignore it, because chart repository
				// is flat and cannot contain nested directories.
				continue
			}

			if !strings.HasSuffix(key, ".tgz") {
				// Ignore any file that isn't a chart
				// This could include index.yaml
				// or any other kind of file that might be in the repo
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

			reindexItem := ChartInfo{Filename: key}

			serializedChartMeta, hasMeta := metaOut.Metadata[strings.Title(metaChartMetadata)]
			chartDigest, hasDigest := metaOut.Metadata[strings.Title(metaChartDigest)]
			if !hasMeta || !hasDigest {
				// TODO: This is deprecated. Remove in the next major release? Or not?
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

// ChartInfo contains info about particular chart.
type ChartInfo struct {
	Meta     *chart.Metadata
	Filename string
	Hash     string
}

// FetchRaw downloads the object from URI and returns it in the form of byte slice.
// uri must be in the form of s3 protocol: s3://bucket-name/key[...].
func (s *Storage) FetchRaw(ctx context.Context, uri string) ([]byte, error) {
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

// Exists returns true if an object exists in the storage.
func (s *Storage) Exists(ctx context.Context, uri string) (bool, error) {
	if _, err := s.GetMetadata(ctx, uri); err != nil {
		// That's weird that there is no NotFound constant in aws sdk.
		if ae, ok := err.(awserr.Error); ok && ae.Code() == "NotFound" {
			return false, nil
		}
		return false, errors.Wrap(err, "head s3 object")
	}

	return true, nil
}

// GetMetadata returns metadata associated with the object in storage
func (s *Storage) GetMetadata(ctx context.Context, uri string) (map[string]string, error) {
	bucket, key, err := parseURI(uri)
	if err != nil {
		return nil, err
	}

	result, err := s3.New(s.session).HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, err
	}

	return aws.StringValueMap(result.Metadata), nil
}

// PutChart puts the chart file to the storage.
// uri must be in the form of s3 protocol: s3://bucket-name/key[...].
func (s *Storage) PutChart(ctx context.Context, uri string, r io.Reader, chartMeta, acl string, chartDigest string, contentType string) (string, error) {
	bucket, key, err := parseURI(uri)
	if err != nil {
		return "", err
	}
	result, err := s3manager.NewUploader(s.session).UploadWithContext(
		ctx,
		&s3manager.UploadInput{
			Bucket:               aws.String(bucket),
			Key:                  aws.String(key),
			ACL:                  aws.String(acl),
			ContentType:          aws.String(contentType),
			ServerSideEncryption: getSSE(),
			Body:                 r,
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

// PutIndex puts the index file to the storage.
// uri must be in the form of s3 protocol: s3://bucket-name/key[...].
func (s *Storage) PutIndex(ctx context.Context, uri string, publishURI string, acl string, r io.Reader) error {
	if strings.HasPrefix(uri, "index.yaml") {
		return errors.New("uri must not contain \"index.yaml\" suffix, it appends automatically")
	}
	uri += "/index.yaml"

	bucket, key, err := parseURI(uri)
	if err != nil {
		return err
	}
	_, err = s3manager.NewUploader(s.session).UploadWithContext(
		ctx,
		&s3manager.UploadInput{
			Bucket:               aws.String(bucket),
			Key:                  aws.String(key),
			ACL:                  aws.String(acl),
			ServerSideEncryption: getSSE(),
			Body:                 r,
			Metadata: map[string]*string{
				MetaPublishURI: aws.String(publishURI),
			},
		})
	if err != nil {
		return errors.Wrap(err, "upload index to S3 bucket")
	}

	return nil
}

// Delete deletes the object by uri.
// uri must be in the form of s3 protocol: s3://bucket-name/key[...].
func (s *Storage) Delete(ctx context.Context, uri string) error {
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
	// metaChartMetadata is a s3 object metadata key that represents chart metadata.
	metaChartMetadata = "chart-metadata"

	// metaChartDigest is a s3 object metadata key that represents chart digest.
	metaChartDigest = "chart-digest"

	// MetaPublishURI s3 object metadata key that stores the non-s3 URI to publish
	MetaPublishURI = "helm-s3-publish-uri"
)
