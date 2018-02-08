package awsutil

import (
	"os"
	"testing"
)

func TestSessionWithCustomEndpoint(t *testing.T) {
	os.Setenv("AWS_ENDPOINT", "foobar:1234")
	os.Setenv("AWS_DISABLE_SSL", "true")

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

	os.Unsetenv("AWS_ENDPOINT")
	os.Unsetenv("AWS_DISABLE_SSL")
}
