package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/awsutil"
	"github.com/hypnoglow/helm-s3/internal/helmutil"
)

type deleteAction struct {
	name, version, repoName, acl string
}

func (act deleteAction) Run(ctx context.Context) error {
	repoEntry, err := helmutil.LookupRepoEntry(act.repoName)
	if err != nil {
		return err
	}

	sess, err := awsutil.Session()
	if err != nil {
		return err
	}
	storage := awss3.New(sess)

	// Fetch current index.
	b, err := storage.FetchRaw(ctx, repoEntry.IndexURL())
	if err != nil {
		return errors.WithMessage(err, "fetch current repo index")
	}

	idx := helmutil.NewIndex()
	if err := idx.UnmarshalBinary(b); err != nil {
		return errors.WithMessage(err, "load index from downloaded file")
	}

	// Update index.

	url, err := idx.Delete(act.name, act.version)
	if err != nil {
		return err
	}

	idxReader, err := idx.Reader()
	if err != nil {
		return errors.Wrap(err, "get index reader")
	}

	// Delete the file from S3 and replace index file.

	if url != "" {
		if err := storage.Delete(ctx, url); err != nil {
			return errors.WithMessage(err, "delete chart file from s3")
		}
	}

	if err := storage.PutIndex(ctx, repoEntry.URL(), act.acl, idxReader); err != nil {
		return errors.WithMessage(err, "upload new index to s3")
	}

	if err := idx.WriteFile(repoEntry.CacheFile(), 0644); err != nil {
		return errors.WithMessage(err, "update local index")
	}

	return nil
}
