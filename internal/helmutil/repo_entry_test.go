package helmutil

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/cli"
	repo3 "helm.sh/helm/v3/pkg/repo"
	repo2 "k8s.io/helm/pkg/repo"
)

func TestLookupRepoEntry(t *testing.T) {
	testCases := map[string]struct {
		setup         func() func()
		name          string
		expectedEntry RepoEntry
		assertError   assert.ErrorAssertionFunc
	}{
		"helm v2": {
			setup: func() func() {
				helm2LoadRepoFile = func(path string) (file *repo2.RepoFile, e error) {
					return &repo2.RepoFile{
						Repositories: []*repo2.Entry{
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
				return mockEnv(t, "TILLER_HOST", "1")
			},
			name: "my-charts",
			expectedEntry: RepoEntryV2{
				entry: &repo2.Entry{
					Name: "my-charts",
					URL:  "s3://my-charts",
				},
			},
			assertError: assert.NoError,
		},
		"helm v2 repo file not found": {
			setup: func() func() {
				helm2LoadRepoFile = func(path string) (file *repo2.RepoFile, e error) {
					_, err := os.Stat("foobarbaz")
					return nil, err
				}
				return mockEnv(t, "TILLER_HOST", "1")
			},
			name:          "my-charts",
			expectedEntry: RepoEntryV2{},
			assertError: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, fs.ErrNotExist)
			},
		},
		"helm v3": {
			setup: func() func() {
				helm3LoadRepoFile = func(path string) (file *repo3.File, e error) {
					return &repo3.File{
						Repositories: []*repo3.Entry{
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
				helm3Detected = func() bool {
					return true
				}
				return func() {}
			},
			name: "my-charts",
			expectedEntry: RepoEntryV3{
				entry: &repo3.Entry{
					Name: "my-charts",
					URL:  "s3://my-charts",
				},
			},
			assertError: assert.NoError,
		},
		"helm v3 repo file not found": {
			setup: func() func() {
				helm3LoadRepoFile = func(path string) (file *repo3.File, e error) {
					_, err := os.Stat("foobarbaz")
					return nil, err
				}
				helm3Env = cli.New()
				helm3Detected = func() bool {
					return true
				}
				return func() {}
			},
			name:          "my-charts",
			expectedEntry: RepoEntryV3{},
			assertError: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, fs.ErrNotExist)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			teardown := tc.setup()
			defer teardown()

			entry, err := LookupRepoEntry(tc.name)
			tc.assertError(t, err)
			assert.Equal(t, tc.expectedEntry, entry)
		})
	}
}
