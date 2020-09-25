package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/awsutil"
	"github.com/hypnoglow/helm-s3/internal/helmutil"
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
	relative       bool
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

	chart, err := helmutil.LoadChart(fname)
	if err != nil {
		return err
	}

	repoEntry, err := helmutil.LookupRepoEntry(act.repoName)
	if err != nil {
		return err
	}

	if cachedIndex, err := helmutil.LoadIndex(repoEntry.CacheFile()); err == nil {
		// if cached index exists, check if the same chart version exists in it.
		if cachedIndex.Has(chart.Name(), chart.Version()) {
			if act.ignoreIfExists {
				return nil
			}
			if !act.force {
				return ErrChartExists
			}

			// fallthrough on --force
		}
	}

	hash, err := helmutil.DigestFile(fname)
	if err != nil {
		return errors.WithMessage(err, "get chart digest")
	}

	fchart, err := os.Open(fname)
	if err != nil {
		return errors.Wrap(err, "open chart file")
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
		chartMetaJSON, err := chart.Metadata().MarshalJSON()
		if err != nil {
			return err
		}
		if _, err := storage.PutChart(ctx, repoEntry.URL()+"/"+fname, fchart, string(chartMetaJSON), act.acl, hash, act.contentType); err != nil {
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

	idx := helmutil.NewIndex()
	if err := idx.UnmarshalBinary(b); err != nil {
		return errors.WithMessage(err, "load index from downloaded file")
	}
	baseURL := repoEntry.URL()
	if act.relative {
		baseURL = ""
	}
	if err := idx.AddOrReplace(chart.Metadata().Value(), fname, baseURL, hash); err != nil {
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
