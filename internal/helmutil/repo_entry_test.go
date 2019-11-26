package helmutil

import (
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
		expectError   bool
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
			expectError: false,
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
			expectError: false,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			teardown := tc.setup()
			defer teardown()

			entry, err := LookupRepoEntry(tc.name)
			assertError(t, err, tc.expectError)
			assert.Equal(t, tc.expectedEntry, entry)
		})
	}
}
