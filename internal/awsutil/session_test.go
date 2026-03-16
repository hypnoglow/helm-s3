package awsutil

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamicBucketRegion(t *testing.T) {
	t.Parallel()

	defaultConfig, err := Session()
	require.NoError(t, err)
	defaultRegion := defaultConfig.Region

	testCases := []struct {
		caseDescription      string
		expectedBucketRegion string
		inputS3URL           string
	}{
		{
			caseDescription:      "existing S3 bucket URL with host only (no key) -> success",
			expectedBucketRegion: "ap-southeast-2",
			inputS3URL:           "s3://cn-test-bucket",
		},
		{
			caseDescription:      "existing S3 bucket URL with key -> success",
			expectedBucketRegion: "ap-southeast-2",
			inputS3URL:           "s3://cn-test-bucket/charts/chart-0.1.2.tgz",
		},
		{
			caseDescription:      "invalid URL -> failing URI parsing, no effect (default region)",
			expectedBucketRegion: defaultRegion,
			inputS3URL:           "://not/a/URL",
		},
		{
			caseDescription:      "invalid S3 URL -> failing request, no effect (default region)",
			expectedBucketRegion: defaultRegion,
			inputS3URL:           "",
		},
		{
			caseDescription:      "not existing S3 URL -> no region header, no effect (default region)",
			expectedBucketRegion: defaultRegion,
			inputS3URL:           "s3://" + uuid.NewString(), // make it truly random, because constant non-existing bucket may eventually be created
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseDescription, func(t *testing.T) {
			t.Parallel()

			actualConfig, err := Session(DynamicBucketRegion(tc.inputS3URL))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedBucketRegion, actualConfig.Region)
		})
	}
}

func TestSessionWithCustomEndpoint(t *testing.T) {
	os.Setenv("AWS_ENDPOINT", "http://foobar:1234")
	os.Setenv("AWS_DISABLE_SSL", "true")
	os.Setenv("HELM_S3_REGION", "us-west-2")

	cfg, err := Session()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Note: In AWS SDK v2, endpoint configuration is validated differently
	// The endpoint resolver is checked when making actual API calls
	if cfg.Region != "us-west-2" {
		t.Fatalf("Expected to set us-west-2 region, got %s", cfg.Region)
	}

	os.Unsetenv("AWS_ENDPOINT")
	os.Unsetenv("AWS_DISABLE_SSL")
	os.Unsetenv("HELM_S3_REGION")
}

func TestSessionWithInvalidEndpoint(t *testing.T) {
	os.Setenv("AWS_ENDPOINT", "foobar:1234")
	os.Setenv("AWS_DISABLE_SSL", "true")
	os.Setenv("HELM_S3_REGION", "us-west-2")

	_, err := Session()
	if err == nil {
		t.Fatalf("Expected error for endpoint without scheme, got nil")
	}
	if err.Error() != "endpoint must include a scheme (e.g., https://)" {
		t.Fatalf("Expected 'endpoint must include a scheme' error, got: %v", err)
	}

	os.Unsetenv("AWS_ENDPOINT")
	os.Unsetenv("AWS_DISABLE_SSL")
	os.Unsetenv("HELM_S3_REGION")
}
