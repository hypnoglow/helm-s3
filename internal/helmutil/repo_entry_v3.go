package helmutil

import (
	"path/filepath"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/repo"
)

// RepoEntryV2 implements RepoEntry in Helm v3.
type RepoEntryV3 struct {
	entry *repo.Entry
}

func (r RepoEntryV3) URL() string {
	return r.entry.URL
}

func (r RepoEntryV3) IndexURL() string {
	return indexFile(r.entry.URL)
}

func (r RepoEntryV3) CacheFile() string {
	return filepath.Join(cacheDirPathV3(), repoCacheFileName(r.entry.Name))
}

func lookupV3(name string) (RepoEntryV3, error) {
	repoFile, err := helm3LoadRepoFile(repoFilePathV3())
	if err != nil {
		return RepoEntryV3{}, errors.Wrap(err, "load repo file")
	}

	entry := repoFile.Get(name)
	if entry == nil {
		return RepoEntryV3{}, errors.Errorf("repo with name %s not found, try `helm repo add %s <uri>`", name, name)
	}

	return RepoEntryV3{entry: entry}, nil
}
