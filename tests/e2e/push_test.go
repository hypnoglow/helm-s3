package e2e

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/repo"

	"github.com/hypnoglow/helm-s3/internal/helmutil"
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

	tmpdir := setupTempDir(t)

	// Fetch the repo index before push to get generated time.

	indexFile := filepath.Join(tmpdir, "index.yaml")

	err := mc.FGetObject(repoName, repoDir+"/index.yaml", indexFile, minio.GetObjectOptions{})
	require.NoError(t, err)

	idx, err := repo.LoadIndexFile(indexFile)
	require.NoError(t, err)

	assert.NotEmpty(t, idx.Generated)

	idxGeneratedTimeBeforePush := idx.Generated

	// Push chart.

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push %s %s", chartFilepath, repoName))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

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

	expected = "The chart already exists in the repository and cannot be overwritten without an explicit intent."
	assert.Contains(t, stderr.String(), expected)

	// Fetch the repo index again and check that generated time was updated.

	err = mc.FGetObject(repoName, repoDir+"/index.yaml", indexFile, minio.GetObjectOptions{})
	require.NoError(t, err)

	idx, err = repo.LoadIndexFile(indexFile)
	require.NoError(t, err)

	assert.Greater(t, idx.Generated, idxGeneratedTimeBeforePush)
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
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

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
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

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
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

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
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

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
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

	// Check that chart was actually pushed and remember last modification time.

	obj, err := mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, chartObjectName, obj.Key)

	lastModified := obj.LastModified

	// Push chart again with --ignore-if-exists.

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 push %s %s --ignore-if-exists", chartFilepath, repoName))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "The chart already exists in the repository, keep existing chart and ignore push.")

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

	tmpdir := setupTempDir(t)

	// Push chart.

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push --relative %s %s", chartFilepath, repoName))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

	// Fetch the repo index and check that chart uri is relative.

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

func TestPushProvenance(t *testing.T) {
	t.Log("Test push action for chart with .prov file")

	const (
		repoName        = "test-push-provenance"
		repoDir         = "charts"
		chartName       = "foo"
		chartVersion    = "1.3.1"
		chartFilename   = "foo-1.3.1.tgz"
		provFilename    = chartFilename + ".prov"
		chartFilepath   = "testdata/" + chartFilename
		chartObjectName = repoDir + "/" + chartFilename
		provObjectName  = chartObjectName + ".prov"
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	tmpdir := setupTempDir(t)

	gnupgHome, err := filepath.Abs("testdata/gnupg")
	require.NoError(t, err)
	t.Setenv("GNUPGHOME", gnupgHome)

	// Push chart.

	cmd := fmt.Sprintf("helm s3 push %s %s", chartFilepath, repoName)
	stdout, stderr, err := runCommand(cmd)
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

	// Check that chart was actually pushed.

	obj, err := mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, chartObjectName, obj.Key)

	// Check that .prov file was actually pushed.

	obj, err = mc.StatObject(repoName, provObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, provObjectName, obj.Key)

	// Check that chart can be fetched and verified.

	cmd = fmt.Sprintf("helm fetch %s/%s --version %s --destination %s --verify", repoName, chartName, chartVersion, tmpdir)
	stdout, stderr, err = runCommand(cmd)
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)

	if helmutil.IsHelm3() {
		assert.Equal(t, `Signed by: Test Key (helm-s3) <test@example.org>
Using Key With Fingerprint: 362A31CD7E4CF7E098605C67491EB34640FC8895
Chart Hash Verified: sha256:be99ea067f7cccf2516cb1e2e43a569f9032d8cde2f28efef200c66662e0b494
`, stdout.String())
	} else {
		assert.Contains(t, stdout.String(), "sha256:be99ea067f7cccf2516cb1e2e43a569f9032d8cde2f28efef200c66662e0b494")
	}

	assert.FileExists(t, filepath.Join(tmpdir, chartFilename))
	assert.FileExists(t, filepath.Join(tmpdir, provFilename))
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
