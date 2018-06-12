package e2e

import (
	"fmt"
	"testing"

	"github.com/minio/minio-go"
)

func TestPush(t *testing.T) {
	t.Log("Test basic push action")

	name := "test-push"
	dir := "charts"
	setupRepo(t, name, dir)
	defer teardownRepo(t, name)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push testdata/foo-1.2.3.tgz %s", name))
	if err := cmd.Run(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if stdout.String() != "" {
		t.Errorf("Expected stdout to be empty, but got %q", stdout.String())
	}

	if stderr.String() != "" {
		t.Errorf("Expected stderr to be empty, but got %q", stderr.String())
	}

	key := dir + "/foo-1.2.3.tgz"

	// Check that chart was actually pushed
	obj, err := mc.StatObject(name, key, minio.StatObjectOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if obj.Key != key {
		t.Errorf("Expected key to be %q but got %q", key, obj.Key)
	}

	// cleanup

	if err = mc.RemoveObject(name, key); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPushDryRun(t *testing.T) {
	t.Log("Test push action with --dry-run flag")

	name := "test-push-dry-run"
	dir := "charts"
	setupRepo(t, name, dir)
	defer teardownRepo(t, name)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push testdata/foo-1.2.3.tgz %s --dry-run", name))
	if err := cmd.Run(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if stdout.String() != "" {
		t.Errorf("Expected stdout to be empty, but got %q", stdout.String())
	}

	if stderr.String() != "" {
		t.Errorf("Expected stderr to be empty, but got %q", stderr.String())
	}

	// Check that actually nothing got pushed

	_, err := mc.StatObject(name, dir+"/foo-1.2.3.tgz", minio.StatObjectOptions{})
	if minio.ToErrorResponse(err).Code != "NoSuchKey" {
		t.Fatalf("Expected chart not to be pushed")
	}
}
