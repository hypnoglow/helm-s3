package e2e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/repo"
)

const (
	defaultChartsContentType = "application/gzip"
)

func TestPush(t *testing.T) {
	t.Log("Test basic push action")

	const (
		repoName        = "test-push"
		repoDir         = "charts"
		chartName       = "foo"
		chartVersion    = "1.2.3"
		chartFilename   = "foo-1.2.3.tgz"
		chartFilepath   = "testdata/" + chartFilename
		chartObjectName = repoDir + "/" + chartFilename
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push %s %s", chartFilepath, repoName))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, stdout, stderr)

	// Check that chart was actually pushed.

	obj, err := mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, chartObjectName, obj.Key)

	// Check that chart has proper content type.

	assert.Equal(t, defaultChartsContentType, obj.ContentType)

	// Check that pushed chart can be searched.

	cmd, stdout, stderr = command(makeSearchCommand(repoName, chartName))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)

	expected := `test-push/foo	1.2.3        	1.2.3      	A Helm chart for Kubernetes`
	assert.Contains(t, stdout.String(), expected)

	// Check that pushed chart can be fetched.

	tmpdir, err := ioutil.TempDir("", t.Name())
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	cmd, stdout, stderr = command(fmt.Sprintf("helm fetch %s/%s --version %s --destination %s", repoName, chartName, chartVersion, tmpdir))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, stdout, stderr)
	assert.FileExists(t, filepath.Join(tmpdir, chartFilename))

	// Check that pushing the same chart again fails.

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 push %s %s", chartFilepath, repoName))
	err = cmd.Run()
	assert.Error(t, err)
	assertEmptyOutput(t, stdout, nil)

	expected = `The chart already exists in the repository and cannot be overwritten without an explicit intent. If you want to replace existing chart, use --force flag`
	assert.Contains(t, stderr.String(), expected)
}

func TestPushContentType(t *testing.T) {
	t.Log("Test push action with --content-type flag")

	const (
		repoName        = "test-push-content-type"
		repoDir         = "charts"
		chartFilename   = "foo-1.2.3.tgz"
		chartFilepath   = "testdata/" + chartFilename
		chartObjectName = repoDir + "/" + chartFilename

		contentType = defaultChartsContentType + "-test"
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push --content-type=%s %s %s", contentType, chartFilepath, repoName))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, stdout, stderr)

	// Check that chart was actually pushed.

	obj, err := mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, chartObjectName, obj.Key)

	// Check that chart has proper content type.

	assert.Equal(t, contentType, obj.ContentType)
}

func TestPushDryRun(t *testing.T) {
	t.Log("Test push action with --dry-run flag")

	const (
		repoName        = "test-push-dry-run"
		repoDir         = "charts"
		chartFilename   = "foo-1.2.3.tgz"
		chartFilepath   = "testdata/" + chartFilename
		chartObjectName = repoDir + "/" + chartFilename
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push %s %s --dry-run", chartFilepath, repoName))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, stdout, stderr)

	// Check that actually nothing got pushed.

	_, err = mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.Equal(t, "NoSuchKey", minio.ToErrorResponse(err).Code)
}

func TestPushForce(t *testing.T) {
	t.Log("Test push action with --force flag")

	const (
		repoName        = "test-push-force"
		repoDir         = "charts"
		chartFilename   = "foo-1.2.3.tgz"
		chartFilepath   = "testdata/" + chartFilename
		chartObjectName = repoDir + "/" + chartFilename
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push %s %s", chartFilepath, repoName))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, stdout, stderr)

	// Check that chart was actually pushed and remember last modification time.

	obj, err := mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, chartObjectName, obj.Key)

	lastModified := obj.LastModified

	// Push chart again with --force.

	time.Sleep(time.Second)

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 push %s %s --force", chartFilepath, repoName))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, stdout, stderr)

	// Check that chart was overwritten.

	obj, err = mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.True(t, obj.LastModified.After(lastModified), "Expected %s less than %s", lastModified.String(), obj.LastModified.String())
}

func TestPushIgnoreIfExists(t *testing.T) {
	t.Log("Test push action with --ignore-if-exists flag")

	const (
		repoName        = "test-push-ignore-if-exists"
		repoDir         = "charts"
		chartFilename   = "foo-1.2.3.tgz"
		chartFilepath   = "testdata/" + chartFilename
		chartObjectName = repoDir + "/" + chartFilename
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push %s %s", chartFilepath, repoName))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, stdout, stderr)

	// Check that chart was actually pushed and remember last modification time.

	obj, err := mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, chartObjectName, obj.Key)

	lastModified := obj.LastModified

	// Push chart again with --ignore-if-exists.

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 push %s %s --ignore-if-exists", chartFilepath, repoName))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, stdout, stderr)

	// Check that chart was not overwritten.

	obj, err = mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, lastModified, obj.LastModified)
}

func TestPushForceAndIgnoreIfExists(t *testing.T) {
	t.Log("Test push action with both --force and --ignore-if-exists flags")

	const (
		repoName      = "test-push-force-and-ignore-if-exists"
		repoDir       = "charts"
		chartFilename = "foo-1.2.3.tgz"
		chartFilepath = "testdata/" + chartFilename
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push %s %s --force --ignore-if-exists", chartFilepath, repoName))
	err := cmd.Run()
	assert.Error(t, err)
	assertEmptyOutput(t, stdout, nil)

	expectedErrorMessage := "The --force and --ignore-if-exists flags are mutually exclusive and cannot be specified together."
	if !strings.HasPrefix(stderr.String(), expectedErrorMessage) {
		t.Errorf("Expected stderr to begin with %q, but got %q", expectedErrorMessage, stderr.String())
	}
}

func TestPushRelative(t *testing.T) {
	t.Log("Test push action with --relative flag")

	const (
		repoName      = "test-push-relative"
		repoDir       = "charts"
		chartName     = "foo"
		chartVersion  = "1.2.3"
		chartFilename = "foo-1.2.3.tgz"
		chartFilepath = "testdata/" + chartFilename
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push --relative %s %s", chartFilepath, repoName))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, stdout, stderr)

	// Fetch the repo index and check that chart uri is relative.

	tmpdir, err := ioutil.TempDir("", t.Name())
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	indexFile := filepath.Join(tmpdir, "index.yaml")

	err = mc.FGetObject(repoName, repoDir+"/index.yaml", indexFile, minio.GetObjectOptions{})
	require.NoError(t, err)

	idx, err := repo.LoadIndexFile(indexFile)
	require.NoError(t, err)

	cv, err := idx.Get(chartName, chartVersion)
	require.NoError(t, err)

	expected := []string{chartFilename}
	if diff := cmp.Diff(expected, cv.URLs); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}

	// Check that chart can be successfully fetched.

	cmd, stdout, stderr = command(fmt.Sprintf("helm fetch %s/%s --version %s --destination %s", repoName, chartName, chartVersion, tmpdir))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, stdout, stderr)
	assert.FileExists(t, filepath.Join(tmpdir, chartFilename))
}

func assertEmptyOutput(t *testing.T, stdout, stderr *bytes.Buffer) {
	t.Helper()

	if stdout != nil {
		assert.Empty(t, stdout.String(), "Expected stdout to be empty")
	}
	if stderr != nil {
		assert.Empty(t, stderr.String(), "Expected stderr to be empty")
	}
}
