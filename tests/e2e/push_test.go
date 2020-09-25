package e2e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/minio/minio-go/v6"
	"helm.sh/helm/v3/pkg/repo"
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
	assertEmptyOutput(t, stdout, stderr)

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
	assertEmptyOutput(t, stdout, stderr)

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
	assertEmptyOutput(t, stdout, stderr)

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
	assertEmptyOutput(t, stdout, stderr)

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
	assertEmptyOutput(t, stdout, stderr)

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
	assertEmptyOutput(t, stdout, stderr)

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
	assertEmptyOutput(t, stdout, nil)

	expectedErrorMessage := "The --force and --ignore-if-exists flags are mutually exclusive and cannot be specified together."
	if !strings.HasPrefix(stderr.String(), expectedErrorMessage) {
		t.Errorf("Expected stderr to begin with %q, but got %q", expectedErrorMessage, stderr.String())
	}
}

func TestPushRelative(t *testing.T) {
	t.Log("Test push action with --relative flag")

	name := "test-push-relative"
	dir := "charts"
	chartName := "foo"
	chartVer := "1.2.3"
	filename := fmt.Sprintf("%s-%s.tgz", chartName, chartVer)

	setupRepo(t, name, dir)
	defer teardownRepo(t, name)

	// set a cleanup in beforehand
	defer removeObject(t, name, dir+"/"+filename)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push --relative testdata/%s %s", filename, name))
	if err := cmd.Run(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	assertEmptyOutput(t, stdout, stderr)

	tmpdir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	indexFile := filepath.Join(tmpdir, "index.yaml")

	if err := mc.FGetObject(name, dir+"/index.yaml", indexFile, minio.GetObjectOptions{}); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	idx, err := repo.LoadIndexFile(indexFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	v, err := idx.Get(chartName, chartVer)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{filename}
	if diff := cmp.Diff(expected, v.URLs); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}

	os.Chdir(tmpdir)
	cmd, stdout, stderr = command(fmt.Sprintf("helm fetch %s/%s --version %s", name, chartName, chartVer))
	if err := cmd.Run(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	assertEmptyOutput(t, stdout, stderr)
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

func assertEmptyOutput(t *testing.T, stdout, stderr *bytes.Buffer) {
	t.Helper()
	if stdout != nil && stdout.String() != "" {
		t.Errorf("Expected stdout to be empty, but got %q", stdout.String())
	}

	if stderr != nil && stderr.String() != "" {
		t.Errorf("Expected stderr to be empty, but got %q", stderr.String())
	}
}

func removeObject(t *testing.T, name, key string) {
	t.Helper()
	if err := mc.RemoveObject(name, key); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
