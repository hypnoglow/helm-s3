package helmutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/helm/pkg/repo"
)

func TestRepoEntryV2_URL(t *testing.T) {
	testCases := map[string]struct {
		entry RepoEntryV2
		url   string
	}{
		"standard repo": {
			entry: RepoEntryV2{
				entry: &repo.Entry{
					Name: "stable",
					URL:  "https://kubernetes-charts.storage.googleapis.com",
				},
			},
			url: "https://kubernetes-charts.storage.googleapis.com",
		},
		"s3 repo": {
			entry: RepoEntryV2{
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

func TestRepoEntryV2_IndexURL(t *testing.T) {
	testCases := map[string]struct {
		entry RepoEntryV2
		url   string
	}{
		"standard repo": {
			entry: RepoEntryV2{
				entry: &repo.Entry{
					Name: "stable",
					URL:  "https://kubernetes-charts.storage.googleapis.com",
				},
			},
			url: "https://kubernetes-charts.storage.googleapis.com/index.yaml",
		},
		"s3 repo": {
			entry: RepoEntryV2{
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

func TestRepoEntryV2_CacheFile(t *testing.T) {
	// mock helm2 home
	helm2Home = "/home/foo/.helm"

	testCases := map[string]struct {
		entry     RepoEntryV2
		cacheFile string
	}{
		"standard repo with abs cache path": {
			entry: RepoEntryV2{
				entry: &repo.Entry{
					Name:  "stable",
					URL:   "https://kubernetes-charts.storage.googleapis.com",
					Cache: "/home/foo/.helm/repository/cache/stable-index.yaml",
				},
			},
			cacheFile: "/home/foo/.helm/repository/cache/stable-index.yaml",
		},
		"standard repo with rel cache path": {
			entry: RepoEntryV2{
				entry: &repo.Entry{
					Name:  "stable",
					URL:   "https://kubernetes-charts.storage.googleapis.com",
					Cache: "stable-index.yaml",
				},
			},
			cacheFile: "/home/foo/.helm/repository/cache/stable-index.yaml",
		},
		"s3 repo with abs cache path": {
			entry: RepoEntryV2{
				entry: &repo.Entry{
					Name:  "my-charts",
					URL:   "s3://my-charts",
					Cache: "/home/foo/.helm/repository/cache/my-charts-index.yaml",
				},
			},
			cacheFile: "/home/foo/.helm/repository/cache/my-charts-index.yaml",
		},
		"s3 repo with rel cache path": {
			entry: RepoEntryV2{
				entry: &repo.Entry{
					Name:  "my-charts",
					URL:   "s3://my-charts",
					Cache: "my-charts-index.yaml",
				},
			},
			cacheFile: "/home/foo/.helm/repository/cache/my-charts-index.yaml",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.cacheFile, tc.entry.CacheFile())
		})
	}
}

func TestLookupV2(t *testing.T) {
	testCases := map[string]struct {
		setup         func() func()
		name          string
		expectedEntry RepoEntryV2
		expectError   bool
	}{
		"should find existing entry": {
			setup: func() func() {
				helm2LoadRepoFile = func(path string) (file *repo.RepoFile, e error) {
					return &repo.RepoFile{
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
				return func() {}
			},
			name: "my-charts",
			expectedEntry: RepoEntryV2{
				entry: &repo.Entry{
					Name: "my-charts",
					URL:  "s3://my-charts",
				},
			},
			expectError: false,
		},
		"should error on non-existing entry": {
			setup: func() func() {
				helm2LoadRepoFile = func(path string) (file *repo.RepoFile, e error) {
					return &repo.RepoFile{
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
				return func() {}
			},
			name:          "my-super-repo",
			expectedEntry: RepoEntryV2{},
			expectError:   true,
		},
		"should error on repo file load failure": {
			setup: func() func() {
				helm2LoadRepoFile = func(path string) (file *repo.RepoFile, e error) {
					return nil, fmt.Errorf("load failed")
				}
				return func() {}
			},
			name:          "my-charts",
			expectedEntry: RepoEntryV2{},
			expectError:   true,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			teardown := tc.setup()
			defer teardown()

			entry, err := lookupV2(tc.name)
			assertError(t, err, tc.expectError)
			assert.Equal(t, tc.expectedEntry, entry)
		})
	}
}
