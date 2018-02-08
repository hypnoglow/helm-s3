package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/provenance"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/awsutil"
	"github.com/hypnoglow/helm-s3/internal/helmutil"
	"github.com/hypnoglow/helm-s3/internal/index"
)

type pushAction struct {
	chartPath string
	repoName  string
}

func (act pushAction) Run(ctx context.Context) error {
	sess, err := awsutil.Session()
	if err != nil {
		return err
	}
	storage := awss3.New(sess)

	fpath, err := filepath.Abs(act.chartPath)
	if err != nil {
		return errors.WithMessage(err, "get chart abs path")
	}

	dir := filepath.Dir(fpath)
	fname := filepath.Base(fpath)

	if err := os.Chdir(dir); err != nil {
		return errors.Wrapf(err, "change dir to %s", dir)
	}

	// Load chart, calculate required params like hash,
	// and upload the chart right away.

	chart, err := chartutil.LoadFile(fname)
	if err != nil {
		return fmt.Errorf("file %s is not a helm chart archive", fname)
	}

	repoEntry, err := helmutil.LookupRepoEntry(act.repoName)
	if err != nil {
		return err
	}

	hash, err := provenance.DigestFile(fname)
	if err != nil {
		return errors.WithMessage(err, "get chart digest")
	}

	fchart, err := os.Open(fname)
	if err != nil {
		return errors.Wrap(err, "open chart file")
	}

	serializedChartMeta, err := json.Marshal(chart.Metadata)
	if err != nil {
		return errors.Wrap(err, "encode chart metadata to json")
	}

	if _, err := storage.PutChart(ctx, repoEntry.URL+"/"+fname, fchart, string(serializedChartMeta), hash); err != nil {
		return errors.WithMessage(err, "upload chart to s3")
	}

	// The gap between index fetching and uploading should be as small as
	// possible to make the best effort to avoid race conditions.
	// See https://github.com/hypnoglow/helm-s3/issues/18 for more info.

	// Fetch current index, update it and upload it back.

	b, err := storage.FetchRaw(ctx, repoEntry.URL+"/index.yaml")
	if err != nil {
		return errors.WithMessage(err, "fetch current repo index")
	}

	idx := &index.Index{}
	if err := idx.UnmarshalBinary(b); err != nil {
		return errors.WithMessage(err, "load index from downloaded file")
	}

	idx.Add(chart.GetMetadata(), fname, repoEntry.URL, hash)
	idx.SortEntries()

	idxReader, err := idx.Reader()
	if err != nil {
		return errors.WithMessage(err, "get index reader")
	}

	if err := storage.PutIndex(ctx, repoEntry.URL, idxReader); err != nil {
		return errors.WithMessage(err, "upload index to s3")
	}

	if err := idx.WriteFile(repoEntry.Cache, 0644); err != nil {
		return errors.WithMessage(err, "update local index")
	}

	return nil
}
