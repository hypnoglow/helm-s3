package main

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/helmutil"
	"github.com/hypnoglow/helm-s3/pkg/index"
)

const (
	reindexCommandDefaultTimeout = time.Second * 15
)

func runReindex(repoName string) error {
	// Just one big timeout for the whole operation.
	ctx, cancel := context.WithTimeout(context.Background(), reindexCommandDefaultTimeout)
	defer cancel()

	repoEntry, err := helmutil.LookupRepoEntry(repoName)
	if err != nil {
		return err
	}

	storage := awss3.New()

	items, errs := storage.Traverse(ctx, repoEntry.URL)

	builtIndex := make(chan *index.Index, 1)
	go func() {
		idx := index.New()
		for item := range items {
			idx.Add(item.Meta, item.Filename, repoEntry.URL, item.Hash)
		}
		idx.SortEntries()

		builtIndex <- idx
	}()

	for err = range errs {
		return errors.Wrap(err, "traverse the chart repository")
	}

	idx := <-builtIndex

	r, err := idx.Reader()
	if err != nil {
		return errors.Wrap(err, "get index reader")
	}

	if err := storage.PutIndex(ctx, repoEntry.URL, r); err != nil {
		return errors.Wrap(err, "upload index to the repository")
	}

	return nil
}
