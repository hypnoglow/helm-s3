package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/awsutil"
	"github.com/hypnoglow/helm-s3/pkg/helmutil"
	"github.com/hypnoglow/helm-s3/pkg/index"
)

type reindexAction struct {
	repoName string
}

func (act reindexAction) Run(ctx context.Context) error {
	repoEntry, err := helmutil.LookupRepoEntry(act.repoName)
	if err != nil {
		return err
	}

	sess, err := awsutil.Session()
	if err != nil {
		return err
	}
	storage := awss3.New(sess)

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
