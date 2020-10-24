package awsutil

import (
	"os"
	"testing"
)

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
