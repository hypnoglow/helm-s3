package helmutil

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/repo"
)

// RepoEntryV3 implements RepoEntry in Helm v3.
type RepoEntryV3 struct {
	entry *repo.Entry
}

func (r RepoEntryV3) Name() string {
	return r.entry.Name
}

func (r RepoEntryV3) URL() string {
	return r.entry.URL
}

func (r RepoEntryV3) IndexURL() string {
	return IndexFileURL(r.entry.URL)
}

func (r RepoEntryV3) CacheFile() string {
	return filepath.Join(cacheDirPathV3(), repoCacheFileName(r.entry.Name))
}

func lookupV3(name string) (RepoEntryV3, error) {
	repoFile, err := helm3LoadRepoFile(repoFilePathV3())
	if err != nil {
		return RepoEntryV3{}, fmt.Errorf("load repo file: %w", err)
	}

	entry := repoFile.Get(name)
	if entry == nil {
		return RepoEntryV3{}, errors.Errorf("repo with name %s not found, try `helm repo add %s <uri>`", name, name)
	}

	return RepoEntryV3{entry: entry}, nil
}

func lookupByURLV3(url string) (RepoEntryV3, bool, error) {
	repoFile, err := helm3LoadRepoFile(repoFilePathV3())
	if err != nil {
		return RepoEntryV3{}, false, fmt.Errorf("load repo file: %w", err)
	}

	url = strings.TrimSuffix(url, "/")
	for _, entry := range repoFile.Repositories {
		entryURL := strings.TrimSuffix(entry.URL, "/")
		if url == entryURL {
			return RepoEntryV3{entry: entry}, true, nil
		}
	}

	return RepoEntryV3{}, false, nil
}
