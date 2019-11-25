package helmutil

import (
	"fmt"
	"testing"

	"helm.sh/helm/v3/pkg/cli"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/repo"
)

func TestRepoEntryV3_URL(t *testing.T) {
	testCases := map[string]struct {
		entry RepoEntryV3
		url   string
	}{
		"standard repo": {
			entry: RepoEntryV3{
				entry: &repo.Entry{
					Name: "stable",
					URL:  "https://kubernetes-charts.storage.googleapis.com/",
				},
			},
			url: "https://kubernetes-charts.storage.googleapis.com/",
		},
		"s3 repo": {
			entry: RepoEntryV3{
				entry: &repo.Entry{
					Name: "my-charts",
					URL:  "s3://my-charts",
				},
			},
			url: "s3://my-charts",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.url, tc.entry.URL())
		})
	}
}

func TestRepoEntryV3_IndexURL(t *testing.T) {
	testCases := map[string]struct {
		entry RepoEntryV3
		url   string
	}{
		"standard repo": {
			entry: RepoEntryV3{
				entry: &repo.Entry{
					Name: "stable",
					URL:  "https://kubernetes-charts.storage.googleapis.com/",
				},
			},
			url: "https://kubernetes-charts.storage.googleapis.com/index.yaml",
		},
		"s3 repo": {
			entry: RepoEntryV3{
				entry: &repo.Entry{
					Name: "my-charts",
					URL:  "s3://my-charts",
				},
			},
			url: "s3://my-charts/index.yaml",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.url, tc.entry.IndexURL())
		})
	}
}

func TestRepoEntryV3_CacheFile(t *testing.T) {
	// mock helm3 env
	helm3Env = cli.New()
	helm3Env.RepositoryCache = "/home/foo/.cache/helm/repository"

	testCases := map[string]struct {
		entry     RepoEntryV3
		cacheFile string
	}{
		"standard repo with abs cache path": {
			entry: RepoEntryV3{
				entry: &repo.Entry{
					Name: "stable",
					URL:  "https://kubernetes-charts.storage.googleapis.com",
				},
			},
			cacheFile: "/home/foo/.cache/helm/repository/stable-index.yaml",
		},
		"standard repo with rel cache path": {
			entry: RepoEntryV3{
				entry: &repo.Entry{
					Name: "stable",
					URL:  "https://kubernetes-charts.storage.googleapis.com",
				},
			},
			cacheFile: "/home/foo/.cache/helm/repository/stable-index.yaml",
		},
		"s3 repo with abs cache path": {
			entry: RepoEntryV3{
				entry: &repo.Entry{
					Name: "my-charts",
					URL:  "s3://my-charts",
				},
			},
			cacheFile: "/home/foo/.cache/helm/repository/my-charts-index.yaml",
		},
		"s3 repo with rel cache path": {
			entry: RepoEntryV3{
				entry: &repo.Entry{
					Name: "my-charts",
					URL:  "s3://my-charts",
				},
			},
			cacheFile: "/home/foo/.cache/helm/repository/my-charts-index.yaml",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.cacheFile, tc.entry.CacheFile())
		})
	}
}

func TestLookupV3(t *testing.T) {
	testCases := map[string]struct {
		setup         func() func()
		name          string
		expectedEntry RepoEntryV3
		expectError   bool
	}{
		"should find existing entry": {
			setup: func() func() {
				helm3LoadRepoFile = func(path string) (file *repo.File, e error) {
					return &repo.File{
						Repositories: []*repo.Entry{
							{
								Name: "stable",
								URL:  "https://kubernetes-charts.storage.googleapis.com",
							},
							{
								Name: "my-charts",
								URL:  "s3://my-charts",
							},
						},
					}, nil
				}
				helm3Env = cli.New()
				return func() {}
			},
			name: "my-charts",
			expectedEntry: RepoEntryV3{
				entry: &repo.Entry{
					Name: "my-charts",
					URL:  "s3://my-charts",
				},
			},
			expectError: false,
		},
		"should error on non-existing entry": {
			setup: func() func() {
				helm3LoadRepoFile = func(path string) (file *repo.File, e error) {
					return &repo.File{
						Repositories: []*repo.Entry{
							{
								Name: "stable",
								URL:  "https://kubernetes-charts.storage.googleapis.com",
							},
							{
								Name: "my-charts",
								URL:  "s3://my-charts",
							},
						},
					}, nil
				}
				helm3Env = cli.New()
				return func() {}
			},
			name:          "my-super-repo",
			expectedEntry: RepoEntryV3{},
			expectError:   true,
		},
		"should error on repo file load failure": {
			setup: func() func() {
				helm3LoadRepoFile = func(path string) (file *repo.File, e error) {
					return nil, fmt.Errorf("load failed")
				}
				helm3Env = cli.New()
				return func() {}
			},
			name:          "my-charts",
			expectedEntry: RepoEntryV3{},
			expectError:   true,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			teardown := tc.setup()
			defer teardown()

			entry, err := lookupV3(tc.name)
			assertError(t, err, tc.expectError)
			assert.Equal(t, tc.expectedEntry, entry)
		})
	}
}
