package awss3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/hypnoglow/helm-s3/internal/helmutil"
)

const (
	// selects serverside encryption for bucket.
	awsS3encryption = "AWS_S3_SSE"

	// s3MetadataSoftLimitBytes is application-specific soft limit
	// for the number of bytes in S3 object metadata.
	s3MetadataSoftLimitBytes = 1900
	localBase                = "/tmp/"
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

// Returns desired encryption.
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
func (s *Storage) Traverse(ctx context.Context, repoURI string) ([]ChartInfo, <-chan error) {
	charts := make(chan ChartInfo)
	errs := make(chan error)
	var result []ChartInfo
	go s.traverse(ctx, repoURI, charts, errs)
	// Collect the results and handle errors
	log.Info("collecting results")
	for {
		select {
		case chart, ok := <-charts:
			if !ok {
				charts = nil
			} else {
				result = append(result, chart)
			}
		}

		// Exit the loop when both channels are closed
		if charts == nil {
			break
		}
	}

	log.Info("collected results")

	return result, errs
}

// traverse traverses all charts in the repository.
// It writes an info item about every chart to items, and errors to errs.
// It always closes both channels when returns.
func (s *Storage) traverse(ctx context.Context, repoURI string, items chan<- ChartInfo, errs chan<- error) {
	log.Info("traversing s3 bucket")
	start := time.Now()
	defer close(items)
	defer close(errs)

	bucket, prefixKey, err := parseURI(repoURI)
	if err != nil {
		log.Errorf("parse uri: %s", err)
		return
	}

	client := s3.New(s.session)

	var continuationToken *string
	var wg sync.WaitGroup

	for {
		log.Info("listing objects")
		listOut, err := client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefixKey),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			log.Errorf("list s3 objects: %s", err)
			return
		}

		log.Info("listOut.Contents: ", len(listOut.Contents))

		wg.Add(1)

		go func() {
			defer wg.Done()
			// Process objects in parallel
			for _, obj := range listOut.Contents {
				processS3Object(ctx, client, bucket, obj, items, prefixKey)
				log.Info("processing object: ", *obj.Key)
			}
		}()

		// Decide if need to load more objects.
		if listOut.NextContinuationToken == nil {
			log.Info("all objects processed")
			break
		}
		continuationToken = listOut.NextContinuationToken
	}
	wg.Wait()

	log.Info("traverse took: ", time.Since(start))
}

func processS3Object(ctx context.Context, client *s3.S3, bucket string, obj *s3.Object, items chan<- ChartInfo, prefixKey string) {
	log.Info("processing object: ", *obj.Key)
	// We need to make object key relative to repo root.
	key := strings.TrimPrefix(*obj.Key, prefixKey)
	// Additionally trim prefix slash if exists, because repos can be:
	// s3://bucket/repo/subdir OR s3://bucket/repo/subdir/
	key = strings.TrimPrefix(key, "/")

	if strings.Contains(key, "/") {
		// This is a subfolder. Ignore it, because chart repository
		// is flat and cannot contain nested directories.
		return
	}

	if !strings.HasSuffix(key, ".tgz") {
		// Ignore any file that isn't a chart
		// This could include index.yaml
		// or any other kind of file that might be in the repo
		return
	}

	metaOut, err := client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    obj.Key,
	})
	if err != nil {
		log.Errorf("head s3 object %q: %s", key, err)
		return
	}

	reindexItem := ChartInfo{Filename: key}

	serializedChartMeta, hasMeta := metaOut.Metadata[strings.Title(metaChartMetadata)]
	chartDigest, hasDigest := metaOut.Metadata[strings.Title(metaChartDigest)]
	if !hasMeta || !hasDigest {
		// Some charts in the repository can have no metadata.
		//
		// This might happen in few cases:
		// - Chart was uploaded manually, not using 'helm s3 push';
		// - Chart was pushed before we started adding metadata to objects;
		// - Chart metadata was too big to add to the S3 object metadata (see issues
		//   https://github.com/hypnoglow/helm-s3/issues/120 and
		//   https://github.com/hypnoglow/helm-s3/issues/112 )
		//
		// In this case we have to download the ch file itself.
		objectOut, err := client.GetObjectWithContext(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    obj.Key,
		})
		if err != nil {
			log.Errorf("get s3 object %q: %s", key, err)
			return
		}

		buf := &bytes.Buffer{}
		tr := io.TeeReader(objectOut.Body, buf)

		ch, err := helmutil.LoadArchive(tr)
		objectOut.Body.Close()
		if err != nil {
			log.Errorf("load archive from s3 object %q: %s", key, err)
			return
		}

		digest, err := helmutil.Digest(buf)
		if err != nil {
			log.Errorf("get chart hash for %q: %s", key, err)
			return
		}

		reindexItem.Meta = ch.Metadata()
		reindexItem.Hash = digest
	} else {
		meta := helmutil.NewChartMetadata()
		if err := meta.UnmarshalJSON([]byte(*serializedChartMeta)); err != nil {
			log.Errorf("unserialize chart meta for %q: %s", key, err)
			return
		}

		reindexItem.Meta = meta
		reindexItem.Hash = *chartDigest
	}

	// process meta and hash
	items <- reindexItem
}

// ChartInfo contains info about particular chart.
type ChartInfo struct {
	Meta     helmutil.ChartMetadata
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
	bucket, key, err := parseURI(uri)
	if err != nil {
		return false, err
	}

	_, err = s3.New(s.session).HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// That's weird that there is no NotFound constant in aws sdk.
		if ae, ok := err.(awserr.Error); ok && ae.Code() == "NotFound" {
			return false, nil
		}
		return false, errors.Wrap(err, "head s3 object")
	}

	return true, nil
}

// PutChart puts the chart file to the storage.
// uri must be in the form of s3 protocol: s3://bucket-name/key[...].
func (s *Storage) PutChart(ctx context.Context, uri string, r io.Reader, chartMeta, acl string, chartDigest string, contentType string, tags string) (string, error) {
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
			Metadata:             assembleObjectMetadata(chartMeta, chartDigest),
			Tagging:              &tags,
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "upload object to s3")
	}

	return result.Location, nil
}

// PutIndex puts the index file to the storage.
// uri must be in the form of s3 protocol: s3://bucket-name/key[...].
func (s *Storage) PutIndex(ctx context.Context, uri string, acl string, r io.Reader) error {
	if strings.HasPrefix(uri, "index.yaml") {
		return errors.New("uri must not contain \"index.yaml\" suffix, it appends automatically")
	}
	uri = helmutil.IndexFileURL(uri)

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
		})
	if err != nil {
		return errors.Wrap(err, "upload index to S3 bucket")
	}

	return nil
}

func (s *Storage) GetIndex(ctx context.Context, uri string, acl string, r io.Reader) error {
	bucket, key, err := parseURI(uri)
	if err != nil {
		return err
	}

	index, err := s3.New(s.session).GetObjectWithContext(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		return errors.Wrap(err, "delete object from s3")
	}

	indexBody, err := io.ReadAll(index.Body)
	if err != nil {
		return errors.Wrap(err, "read index body")
	}

	return yaml.Unmarshal(indexBody, r)
}

// IndexExists returns true if index file exists in the storage for repository
// with the provided uri.
// uri must be in the form of s3 protocol: s3://bucket-name/key[...].
func (s *Storage) IndexExists(ctx context.Context, uri string) (bool, error) {
	if strings.HasPrefix(uri, "index.yaml") {
		return false, errors.New("uri must not contain \"index.yaml\" suffix, it appends automatically")
	}
	uri = helmutil.IndexFileURL(uri)

	return s.Exists(ctx, uri)
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
//   - s3://bucket-name/dir
//   - s3://bucket-name/dir/file.ext
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

// assembleObjectMetadata assembles and returns S3 object metadata.
// May return empty metadata if chart metadata is too big.
//
// The user-defined metadata for the object is limited to 2 KB in size.
// To mitigate the issue with large charts which metadata is more than 2 KB,
// we simply drop it. This affects 'reindex' operation, so that it has to download
// the chart file (GET Request) instead of only fetching its metadata (HEAD request).
func assembleObjectMetadata(chartMeta, chartDigest string) map[string]*string {
	meta := map[string]*string{
		metaChartMetadata: aws.String(chartMeta),
		metaChartDigest:   aws.String(chartDigest),
	}
	if objectMetadataSize(meta) > s3MetadataSoftLimitBytes {
		return nil
	}

	return meta
}

// objectMetadataSize calculates object metadata size as described in https://docs.aws.amazon.com/AmazonS3/latest/dev/UsingMetadata.html
// "The size of user-defined metadata is measured by taking the sum of the number of bytes in the UTF-8 encoding of each key and value.".
func objectMetadataSize(m map[string]*string) int {
	var sum int
	for k, v := range m {
		sum += len([]byte(k))
		if v == nil {
			continue
		}
		sum += len([]byte(*v))
	}
	return sum
}

const (
	// metaChartMetadata is a s3 object metadata key that represents chart metadata.
	metaChartMetadata = "chart-metadata"

	// metaChartDigest is a s3 object metadata key that represents chart digest.
	metaChartDigest = "chart-digest"
)
