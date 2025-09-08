package awsutil

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamicBucketRegion(t *testing.T) {
	t.Parallel()

	defaultSession, err := Session()
	require.NoError(t, err)
	defaultRegion := aws.StringValue(defaultSession.Config.Region)

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

			actualSession, err := Session(DynamicBucketRegion(tc.inputS3URL))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedBucketRegion, aws.StringValue(actualSession.Config.Region))
		})
	}
}

func TestSessionWithCustomEndpoint(t *testing.T) {
	os.Setenv("AWS_ENDPOINT", "foobar:1234")
	os.Setenv("AWS_DISABLE_SSL", "true")
	os.Setenv("HELM_S3_REGION", "us-west-2")

	s, err := Session()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if *s.Config.Endpoint != "foobar:1234" {
		t.Fatalf("Expected endpoint to be foobar:1234")
	}

	if !*s.Config.DisableSSL {
		t.Fatalf("Expected to disable SSL")
	}

	if *s.Config.Region != "us-west-2" {
		t.Fatalf("Expected to set us-west-2 region")
	}
	os.Unsetenv("AWS_ENDPOINT")
	os.Unsetenv("AWS_DISABLE_SSL")
	os.Unsetenv("HELM_S3_REGION")
}

func TestConditionalDynamicBucketRegion(t *testing.T) {
	t.Parallel()

	defaultSession, err := Session()
	require.NoError(t, err)
	defaultRegion := aws.StringValue(defaultSession.Config.Region)

	testCases := []struct {
		caseDescription      string
		envValue             string
		inputS3URL           string
		expectedBucketRegion string
	}{
		{
			caseDescription:      "dynamic region enabled (default) -> should detect region",
			envValue:             "", // not set, should default to enabled
			inputS3URL:           "s3://cn-test-bucket",
			expectedBucketRegion: "ap-southeast-2",
		},
		{
			caseDescription:      "dynamic region explicitly enabled -> should detect region",
			envValue:             "true",
			inputS3URL:           "s3://cn-test-bucket",
			expectedBucketRegion: "ap-southeast-2",
		},
		{
			caseDescription:      "dynamic region disabled -> should use default region",
			envValue:             "false",
			inputS3URL:           "s3://cn-test-bucket",
			expectedBucketRegion: defaultRegion,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseDescription, func(t *testing.T) {
			t.Parallel()

			// Set environment variable for this test case
			if tc.envValue != "" {
				os.Setenv("HELM_S3_DYNAMIC_REGION", tc.envValue)
				defer os.Unsetenv("HELM_S3_DYNAMIC_REGION")
			}

			actualSession, err := Session(ConditionalDynamicBucketRegion(tc.inputS3URL))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedBucketRegion, aws.StringValue(actualSession.Config.Region))
		})
	}
}
