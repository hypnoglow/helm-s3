package main

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/awsutil"
	"github.com/hypnoglow/helm-s3/pkg/helmutil"
	"github.com/hypnoglow/helm-s3/pkg/index"
)

const (
	reindexCommandDefaultTimeput = time.Second * 15
)

func runReindex(repoName string) error {
	// Just one big timeout for the whole operation.
	ctx, cancel := context.WithTimeout(context.Background(), reindexCommandDefaultTimeput)
	defer cancel()

	ctx = ctx

	repoEntry, err := helmutil.LookupRepoEntry(repoName)
	if err != nil {
		return err
	}

	awsConfig, err := awsutil.Config()
	if err != nil {
		return errors.Wrap(err, "get aws config")
	}

	storage := awss3.NewStorage(awsConfig)

	items := make(chan awss3.ChartInfo, 1)
	errs := make(chan error, 1)

	go storage.Traverse(context.TODO(), repoEntry.URL, items, errs)

	builtIndex := make(chan *index.Index, 1)
	go func() {
		idx := index.New()
		for item := range items {
			idx.Add(item.Meta, item.Filename, repoEntry.URL, item.Hash)
		}
		idx.SortEntries()

		builtIndex <- idx
	}()

	for err := range errs {
		return errors.Wrap(err, "traverse the chart repository")
	}

	idx := <-builtIndex

	r, err := idx.Reader()
	if err != nil {
		return errors.Wrap(err, "get index reader")
	}

	if err := storage.PutIndex(context.TODO(), repoEntry.URL, r); err != nil {
		return errors.Wrap(err, "upload index to the repository")
	}

	return nil
}
