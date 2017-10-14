package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/provenance"
	"k8s.io/helm/pkg/repo"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/awsutil"
	"github.com/hypnoglow/helm-s3/pkg/index"
)

func runPush(chartPath string, repoName string) error {
	fpath, err := filepath.Abs(chartPath)
	if err != nil {
		return errors.WithMessage(err, "get chart abs path")
	}

	dir := filepath.Dir(fpath)
	fname := filepath.Base(fpath)
	os.Chdir(dir)

	// Load chart and calculate required params like hash.

	chart, err := chartutil.LoadFile(fname)
	if err != nil {
		return fmt.Errorf("file %s is not a helm chart archive", fname)
	}

	repoURL, err := lookupRepoURL(repoName)
	if err != nil {
		return err
	}

	hash, err := provenance.DigestFile(fname)
	if err != nil {
		return errors.WithMessage(err, "get chart digest")
	}

	// Fetch current index.

	awsConfig, err := awsutil.Config()
	if err != nil {
		return errors.Wrap(err, "get aws config")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	b, err := awss3.FetchRaw(ctx, repoURL+"/index.yaml", awsConfig)
	if err != nil {
		return errors.WithMessage(err, "fetch current repo index")
	}

	idx, err := index.LoadBytes(b)
	if err != nil {
		return errors.WithMessage(err, "load index from downloaded file")
	}

	// Update index.

	idx.Add(chart.GetMetadata(), fname, repoURL, hash)
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
	if _, err := awss3.Upload(ctx, repoURL+"/"+fname, fchart, awsConfig); err != nil {
		return errors.WithMessage(err, "upload chart to s3")
	}
	if _, err := awss3.Upload(ctx, repoURL+"/index.yaml", idxReader, awsConfig); err != nil {
		return errors.WithMessage(err, "upload index to s3")
	}

	return nil
}

func lookupRepoURL(name string) (string, error) {
	h := helmpath.Home(environment.DefaultHelmHome)
	if os.Getenv("HELM_HOME") != "" {
		h = helmpath.Home(os.Getenv("HELM_HOME"))
	}

	repoFile, err := repo.LoadRepositoriesFile(h.RepositoryFile())
	if err != nil {
		return "", errors.Wrap(err, "load repo file")
	}

	for _, r := range repoFile.Repositories {
		if r.Name == name {
			return r.URL, nil
		}
	}

	return "", errors.Errorf("repo with name %s not found, try `helm repo add %s <uri>`", name, name)
}
