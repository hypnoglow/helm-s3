package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/helmutil"
	"github.com/hypnoglow/helm-s3/pkg/index"
)

func runDelete(name, version, repoName string) error {
	repoEntry, err := helmutil.LookupRepoEntry(repoName)
	if err != nil {
		return err
	}

	storage := awss3.New()

	// Fetch current index.

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	b, err := storage.FetchRaw(ctx, repoEntry.URL+"/index.yaml")
	if err != nil {
		return errors.WithMessage(err, "fetch current repo index")
	}

	idx, err := index.LoadBytes(b)
	if err != nil {
		return errors.WithMessage(err, "load index from downloaded file")
	}

	// Update index.

	chartVersion, err := idx.Delete(name, version)
	if err != nil {
		return err
	}

	idxReader, err := idx.Reader()
	if err != nil {
		return errors.Wrap(err, "get index reader")
	}

	// Delete the file from S3 and replace index file.

	if len(chartVersion.URLs) < 1 {
		return fmt.Errorf("chart version index record has no urls")
	}
	uri := chartVersion.URLs[0]

	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout*2)
	defer cancel()

	if err := storage.Delete(ctx, uri); err != nil {
		return errors.WithMessage(err, "delete chart file from s3")
	}
	if err := storage.PutIndex(ctx, repoEntry.URL, idxReader); err != nil {
		return errors.WithMessage(err, "upload new index to s3")
	}

	return nil
}
