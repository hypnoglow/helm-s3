package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/minio/minio-go"
)

const (
	// copied from main
	defaultChartsContentType = "application/gzip"
)

func TestPush(t *testing.T) {
	t.Log("Test basic push action")

	name := "test-push"
	dir := "charts"
	setupRepo(t, name, dir)
	defer teardownRepo(t, name)

	key := dir + "/foo-1.2.3.tgz"

	// set a cleanup in beforehand
	defer removeObject(t, name, key)

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

	// Check that chart was actually pushed
	obj, err := mc.StatObject(name, key, minio.StatObjectOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if obj.Key != key {
		t.Errorf("Expected key to be %q but got %q", key, obj.Key)
	}
}
func TestPushWithContentTypeDefault(t *testing.T) {
	contentType := defaultChartsContentType
	t.Logf("Test basic push action with default Content-Type '%s'", contentType)

	name := "test-push"
	dir := "charts"
	setupRepo(t, name, dir)
	defer teardownRepo(t, name)

	key := dir + "/foo-1.2.3.tgz"

	// set a cleanup in beforehand
	defer removeObject(t, name, key)

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

	assertContentType(t, contentType, name, key)
}

func TestPushWithContentTypeCustom(t *testing.T) {
	contentType := fmt.Sprintf("%s-test", defaultChartsContentType)
	t.Logf("Test basic push action with --content-type='%s'", contentType)

	name := "test-push"
	dir := "charts"
	setupRepo(t, name, dir)
	defer teardownRepo(t, name)

	key := dir + "/foo-1.2.3.tgz"

	// set a cleanup in beforehand
	defer removeObject(t, name, key)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push --content-type=%s testdata/foo-1.2.3.tgz %s", contentType, name))
	if err := cmd.Run(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if stdout.String() != "" {
		t.Errorf("Expected stdout to be empty, but got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Errorf("Expected stderr to be empty, but got %q", stderr.String())
	}

	assertContentType(t, contentType, name, key)
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

func TestPushIgnoreIfExists(t *testing.T) {
	t.Log("Test push action with --ignore-if-exists flag")

	name := "test-push-ignore-if-exists"
	dir := "charts"
	setupRepo(t, name, dir)
	defer teardownRepo(t, name)

	key := dir + "/foo-1.2.3.tgz"

	// set a cleanup in beforehand
	defer removeObject(t, name, key)

	// first, push a chart

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

	// check that chart was actually pushed and remember last modification time

	obj, err := mc.StatObject(name, key, minio.StatObjectOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if obj.Key != key {
		t.Errorf("Expected key to be %q but got %q", key, obj.Key)
	}

	lastModified := obj.LastModified

	// push a chart again with --ignore-if-exists

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 push testdata/foo-1.2.3.tgz %s --ignore-if-exists", name))
	if err := cmd.Run(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if stdout.String() != "" {
		t.Errorf("Expected stdout to be empty, but got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Errorf("Expected stderr to be empty, but got %q", stderr.String())
	}

	// sanity check that chart was not overwritten

	obj, err = mc.StatObject(name, key, minio.StatObjectOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !obj.LastModified.Equal(lastModified) {
		t.Errorf("Expected chart not to be modified")
	}
}

func TestPushForceAndIgnoreIfExists(t *testing.T) {
	t.Log("Test push action with both --force and --ignore-if-exists flags")

	name := "test-push-force-and-ignore-if-exists"
	dir := "charts"
	setupRepo(t, name, dir)
	defer teardownRepo(t, name)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push testdata/foo-1.2.3.tgz %s --force --ignore-if-exists", name))
	if err := cmd.Run(); err == nil {
		t.Errorf("Expected error")
	}

	if stdout.String() != "" {
		t.Errorf("Expected stdout to be empty, but got %q", stdout.String())
	}

	expectedErrorMessage := "The --force and --ignore-if-exists flags are mutually exclusive and cannot be specified together."
	if !strings.HasPrefix(stderr.String(), expectedErrorMessage) {
		t.Errorf("Expected stderr to begin with %q, but got %q", expectedErrorMessage, stderr.String())
	}
}

func assertContentType(t *testing.T, contentType, name, key string) {
	t.Helper()
	obj, err := mc.StatObject(name, key, minio.StatObjectOptions{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if obj.Key != key {
		t.Errorf("Expected key to be %q but got %q", key, obj.Key)
	}
	if obj.ContentType != contentType {
		t.Errorf("Expected ContentType to be %q but got %q", contentType, obj.ContentType)
	}
}

func removeObject(t *testing.T, name, key string) {
	t.Helper()
	if err := mc.RemoveObject(name, key); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
