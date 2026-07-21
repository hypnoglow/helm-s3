package awsutil

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
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
			expectedBucketRegion: "eu-central-1",
			inputS3URL:           "s3://cn-test-bucket",
		},
		{
			caseDescription:      "existing S3 bucket URL with key -> success",
			expectedBucketRegion: "eu-central-1",
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
	t.Setenv("AWS_ENDPOINT", "http://foobar:1234")
	t.Setenv("AWS_DISABLE_SSL", "true")
	t.Setenv("HELM_S3_REGION", "us-west-2")

	cfg, err := Session()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Note: In AWS SDK v2, endpoint configuration is validated differently
	// The endpoint resolver is checked when making actual API calls
	if cfg.Region != "us-west-2" {
		t.Fatalf("Expected to set us-west-2 region, got %s", cfg.Region)
	}
}

func TestDynamicBucketRegionDisabled(t *testing.T) {
	// Stand up an httptest.Server that counts every request it receives, and
	// install a transport that rewrites every outbound request from the AWS
	// SDK to that server. If DynamicBucketRegion respects the disable flag, the
	// server's hit counter must remain at zero. (Without this interception, the
	// test would also pass when the HEAD request was sent but failed for an
	// unrelated reason like DNS or CI sandboxing.)
	var requestsReceived int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestsReceived, 1)
		w.Header().Set("X-Amz-Bucket-Region", "us-west-2")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	origTransport := http.DefaultTransport
	http.DefaultTransport = &rewritingTransport{target: serverURL, base: origTransport}
	defer func() { http.DefaultTransport = origTransport }()

	t.Setenv("HELM_S3_DYNAMIC_REGION_ENABLED", "false")

	defaultCfg, err := Session()
	require.NoError(t, err)
	defaultRegion := defaultCfg.Region

	actualCfg, err := Session(DynamicBucketRegion("s3://cn-test-bucket"))
	require.NoError(t, err)

	assert.Equal(t, defaultRegion, actualCfg.Region)
	assert.Zero(t, atomic.LoadInt32(&requestsReceived),
		"expected no HTTP requests when dynamic region discovery is disabled")
}

func TestSessionWithInvalidEndpoint(t *testing.T) {
	t.Setenv("AWS_ENDPOINT", "foobar:1234")
	t.Setenv("AWS_DISABLE_SSL", "true")
	t.Setenv("HELM_S3_REGION", "us-west-2")

	_, err := Session()
	if err == nil {
		t.Fatalf("Expected error for endpoint without scheme, got nil")
	}
	if err.Error() != "endpoint must include a scheme (e.g., https://)" {
		t.Fatalf("Expected 'endpoint must include a scheme' error, got: %v", err)
	}
}

// rewritingTransport routes every request to target, regardless of the
// request's original scheme/host. Used in tests to point the AWS SDK at an
// httptest.Server.
type rewritingTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (t *rewritingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme = t.target.Scheme
	req.URL.Host = t.target.Host
	req.Host = t.target.Host
	return t.base.RoundTrip(req)
}
