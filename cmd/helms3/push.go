package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/provenance"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/awsutil"
	"github.com/hypnoglow/helm-s3/pkg/helmutil"
	"github.com/hypnoglow/helm-s3/pkg/index"
)

func runPush(chartPath string, repoName string) error {
	fpath, err := filepath.Abs(chartPath)
	if err != nil {
		return errors.WithMessage(err, "get chart abs path")
	}

	dir := filepath.Dir(fpath)
	fname := filepath.Base(fpath)

	if err := os.Chdir(dir); err != nil {
		return errors.Wrapf(err, "change dir to %s", dir)
	}

	awsConfig, err := awsutil.Config()
	if err != nil {
		return errors.Wrap(err, "get aws config")
	}

	storage := awss3.NewStorage(awsConfig)

	// Load chart and calculate required params like hash.

	chart, err := chartutil.LoadFile(fname)
	if err != nil {
		return fmt.Errorf("file %s is not a helm chart archive", fname)
	}

	repoEntry, err := helmutil.LookupRepoEntry(repoName)
	if err != nil {
		return err
	}

	hash, err := provenance.DigestFile(fname)
	if err != nil {
		return errors.WithMessage(err, "get chart digest")
	}

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

	idx.Add(chart.GetMetadata(), fname, repoEntry.URL, hash)
	idx.SortEntries()

	// Finally, upload both chart file and index.

	fchart, err := os.Open(fname)
	if err != nil {
		return errors.Wrap(err, "open chart file")
	}
	idxReader, err := idx.Reader()
	if err != nil {
		return errors.WithMessage(err, "get index reader")
	}

	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout*2)
	defer cancel()

	if _, err := storage.Upload(ctx, repoEntry.URL+"/"+fname, fchart); err != nil {
		return errors.WithMessage(err, "upload chart to s3")
	}
	if _, err := storage.Upload(ctx, repoEntry.URL+"/index.yaml", idxReader); err != nil {
		return errors.WithMessage(err, "upload index to s3")
	}

	return nil
}
