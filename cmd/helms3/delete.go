package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/helm/pkg/repo"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/awsutil"
	"github.com/hypnoglow/helm-s3/internal/helmutil"
	"github.com/hypnoglow/helm-s3/internal/index"
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
	indexURI := repoEntry.URL + "/index.yaml"
	b, err := storage.FetchRaw(ctx, indexURI)
	if err != nil {
		return errors.WithMessage(err, "fetch current repo index")
	}

	idx := &index.Index{}
	if err := idx.UnmarshalBinary(b); err != nil {
		return errors.WithMessage(err, "load index from downloaded file")
	}

	// Update index.

	chartVersion, err := idx.Delete(act.name, act.version)
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

	metadata, err := storage.GetMetadata(ctx, indexURI)
	if err != nil {
		return err
	}

	publishURI := metadata[strings.Title(awss3.MetaPublishURI)]
	uri := fmt.Sprintf("%s/%s-%s.tgz", repoEntry.URL, chartVersion.Metadata.Name, chartVersion.Metadata.Version)

	if err := storage.Delete(ctx, uri); err != nil {
		return errors.WithMessage(err, "delete chart file from s3")
	}
	if err := storage.PutIndex(ctx, repoEntry.URL, publishURI, act.acl, idxReader); err != nil {
		return errors.WithMessage(err, "upload new index to s3")
	}

	localIndexFile, err := repo.LoadIndexFile(repoEntry.Cache)
	if err != nil {
		return err
	}

	localIndex := &index.Index{IndexFile: localIndexFile}

	_, err = localIndex.Delete(act.name, act.version)
	if err != nil {
		return errors.WithMessage(err, "delete chart from local index")
	}

	if err := localIndex.WriteFile(repoEntry.Cache, 0644); err != nil {
		return errors.WithMessage(err, "update local index")
	}

	return nil
}
