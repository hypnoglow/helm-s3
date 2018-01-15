package helmutil

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"

	"github.com/hypnoglow/helm-s3/pkg/index"
)

const (
	envHelmHome = "HELM_HOME"
)

// LookupRepoEntry returns an entry from helm's repositories.yaml file by name.
func LookupRepoEntry(name string) (*repo.Entry, error) {
	h := helmpath.Home(environment.DefaultHelmHome)
	if os.Getenv(envHelmHome) != "" {
		h = helmpath.Home(os.Getenv(envHelmHome))
	}

	repoFile, err := repo.LoadRepositoriesFile(h.RepositoryFile())
	if err != nil {
		return nil, errors.Wrap(err, "load repo file")
	}

	for _, r := range repoFile.Repositories {
		if r.Name == name {
			return r, nil
		}
	}

	return nil, errors.Errorf("repo with name %s not found, try `helm repo add %s <uri>`", name, name)
}

// UpdateLocalIndex rewrites index file for repository named repoName with idx
// contents.
func UpdateLocalIndex(repoName string, idx *index.Index) error {
	entry, err := LookupRepoEntry(repoName)
	if err != nil {
		return err
	}

	b, err := idx.Bytes()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(entry.Cache, b, 0644)
}
