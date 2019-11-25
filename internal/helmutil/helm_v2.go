package helmutil

// This file contains helpers for helm v2.

import (
	"os"

	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
)

// setupHelm2 sets up environment and function bindings for helm v2.
func setupHelm2() {
	helm2Home = resolveHome()
	helm2LoadRepoFile = repo.LoadRepositoriesFile
}

var (
	helm2Home helmpath.Home

	// func that loads helm repo file.
	// Defined for testing purposes.
	helm2LoadRepoFile func(path string) (*repo.RepoFile, error)
)

const (
	envHelmHome = "HELM_HOME"
)

func resolveHome() helmpath.Home {
	h := helmpath.Home(environment.DefaultHelmHome)
	if os.Getenv(envHelmHome) != "" {
		h = helmpath.Home(os.Getenv(envHelmHome))
	}

	return h
}

func repoFilePathV2() string {
	return helm2Home.RepositoryFile()
}

func cacheDirPathV2() string {
	return helm2Home.Cache()
}
