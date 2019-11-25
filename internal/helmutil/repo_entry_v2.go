package helmutil

import (
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/helm/pkg/repo"
)

// RepoEntryV2 implements RepoEntry in Helm v2.
type RepoEntryV2 struct {
	entry *repo.Entry
}

func (r RepoEntryV2) URL() string {
	return r.entry.URL
}

func (r RepoEntryV2) IndexURL() string {
	return indexFile(r.entry.URL)
}

func (r RepoEntryV2) CacheFile() string {
	cache := r.entry.Cache
	if !filepath.IsAbs(cache) {
		cache = filepath.Join(cacheDirPathV2(), cache)
	}
	return cache
}

func lookupV2(name string) (RepoEntryV2, error) {
	repoFile, err := helm2LoadRepoFile(repoFilePathV2())
	if err != nil {
		return RepoEntryV2{}, errors.Wrap(err, "load repo file")
	}

	if entry, ok := repoFile.Get(name); ok {
		return RepoEntryV2{entry: entry}, nil
	}

	return RepoEntryV2{}, errors.Errorf("repo with name %s not found, try `helm repo add %s <uri>`", name, name)
}
