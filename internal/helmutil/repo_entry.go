package helmutil

import (
	"os"

	"github.com/pkg/errors"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
)

const (
	envHelmHome = "HELM_HOME"
)

func getHome() helmpath.Home {
	h := helmpath.Home(environment.DefaultHelmHome)
	if os.Getenv(envHelmHome) != "" {
		h = helmpath.Home(os.Getenv(envHelmHome))
	}

	return h
}

// LookupRepoEntry returns an entry from helm's repositories.yaml file by name.
func LookupRepoEntry(name string) (*repo.Entry, error) {
	h := getHome()

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
