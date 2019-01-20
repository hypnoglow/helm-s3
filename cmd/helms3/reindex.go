package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/awsutil"
	"github.com/hypnoglow/helm-s3/internal/helmutil"
	"github.com/hypnoglow/helm-s3/internal/index"
)

type reindexAction struct {
	repoName   string
	publishURI string
	acl        string
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

	uri := repoEntry.URL
	if act.publishURI != "" {
		uri = act.publishURI
	}

	builtIndex := make(chan *index.Index, 1)
	go func() {
		idx := index.New()
		for item := range items {
			idx.Add(item.Meta, item.Filename, uri, item.Hash)
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

	if err := storage.PutIndex(ctx, repoEntry.URL, act.publishURI, act.acl, r); err != nil {
		return errors.Wrap(err, "upload index to the repository")
	}

	if err := idx.WriteFile(repoEntry.Cache, 0644); err != nil {
		return errors.WithMessage(err, "update local index")
	}

	return nil
}
