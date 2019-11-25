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
	"k8s.io/helm/pkg/repo"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/awsutil"
	"github.com/hypnoglow/helm-s3/internal/helmutil"
	"github.com/hypnoglow/helm-s3/internal/index"
)

var (
	// ErrChartExists signals that chart already exists in the repository
	// and cannot be pushed without --force flag.
	ErrChartExists = errors.New("chart already exists")

	// ErrForceAndIgnoreIfExists signals that the --force and --ignore-if-exists
	// flags cannot be used together.
	ErrForceAndIgnoreIfExists = errors.New("The --force and --ignore-if-exists flags are mutually exclusive and cannot be specified together.")
)

type pushAction struct {
	// required parameters

	chartPath string
	repoName  string

	// optional parameters and flags

	force          bool
	dryRun         bool
	ignoreIfExists bool
	acl            string
	contentType    string
}

func (act pushAction) Run(ctx context.Context) error {
	// Sanity check.
	if act.force && act.ignoreIfExists {
		return ErrForceAndIgnoreIfExists
	}

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

	if cachedIndex, err := repo.LoadIndexFile(repoEntry.CacheFile()); err == nil {
		// if cached index exists, check if the same chart version exists in it.
		if cachedIndex.Has(chart.Metadata.Name, chart.Metadata.Version) {
			if act.ignoreIfExists {
				return nil
			}
			if !act.force {
				return ErrChartExists
			}

			// fallthrough on --force
		}
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

	exists, err := storage.Exists(ctx, repoEntry.URL()+"/"+fname)
	if err != nil {
		return errors.WithMessage(err, "check if chart already exists in the repository")
	}

	if exists {
		if act.ignoreIfExists {
			return nil
		}
		if !act.force {
			return ErrChartExists
		}

		// fallthrough on --force
	}

	if !act.dryRun {
		if _, err := storage.PutChart(ctx, repoEntry.URL()+"/"+fname, fchart, string(serializedChartMeta), act.acl, hash, act.contentType); err != nil {
			return errors.WithMessage(err, "upload chart to s3")
		}
	}

	// The gap between index fetching and uploading should be as small as
	// possible to make the best effort to avoid race conditions.
	// See https://github.com/hypnoglow/helm-s3/issues/18 for more info.

	// Fetch current index, update it and upload it back.

	b, err := storage.FetchRaw(ctx, repoEntry.IndexURL())
	if err != nil {
		return errors.WithMessage(err, "fetch current repo index")
	}

	idx := &index.Index{}
	if err := idx.UnmarshalBinary(b); err != nil {
		return errors.WithMessage(err, "load index from downloaded file")
	}

	if err := idx.AddOrReplace(chart.GetMetadata(), fname, repoEntry.URL(), hash); err != nil {
		return errors.WithMessage(err, "add/replace chart in the index")
	}
	idx.SortEntries()

	idxReader, err := idx.Reader()
	if err != nil {
		return errors.WithMessage(err, "get index reader")
	}

	if !act.dryRun {
		if err := storage.PutIndex(ctx, repoEntry.URL(), act.acl, idxReader); err != nil {
			return errors.WithMessage(err, "upload index to s3")
		}

		if err := idx.WriteFile(repoEntry.CacheFile(), 0644); err != nil {
			return errors.WithMessage(err, "update local index")
		}
	}

	return nil
}
