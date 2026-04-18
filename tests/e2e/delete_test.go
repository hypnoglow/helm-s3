package e2e

import (
	"fmt"
	"testing"

	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	t.Log("Test basic delete action")

	const (
		repoName        = "test-delete"
		repoDir         = "charts"
		chartName       = "foo"
		chartVersion    = "1.2.3"
		chartFilename   = "foo-1.2.3.tgz"
		chartFilepath   = "testdata/" + chartFilename
		chartObjectName = repoDir + "/" + chartFilename
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	// Push chart to be deleted.

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push %s %s", chartFilepath, repoName))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

	// Check that pushed chart exists in the bucket.

	obj, err := mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, chartObjectName, obj.Key)

	// Check that pushed chart can be searched, which means it exists in the index.

	cmd, stdout, stderr = command(makeSearchCommand(repoName, chartName))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)

	expected := `test-delete/foo	1.2.3        	1.2.3      	A Helm chart for Kubernetes`
	assert.Contains(t, stdout.String(), expected)

	// Delete chart.

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 delete %s --version %s %s", chartName, chartVersion, repoName))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully deleted the chart from the repository.")

	// Check that chart was actually deleted from the bucket.

	_, err = mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.Equal(t, "NoSuchKey", minio.ToErrorResponse(err).Code)

	// Check that deleted chart cannot be searched, which means it was deleted from the index.

	cmd, stdout, stderr = command(makeSearchCommand(repoName, chartName))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)

	expected = `No results found`
	assert.Contains(t, stdout.String(), expected)
}

func TestDeleteRelative(t *testing.T) {
	t.Log("Test delete action when chart is pushed with --relative flag")

	const (
		repoName        = "test-delete-relative"
		repoDir         = "charts"
		chartName       = "foo"
		chartVersion    = "1.2.3"
		chartFilename   = "foo-1.2.3.tgz"
		chartFilepath   = "testdata/" + chartFilename
		chartObjectName = repoDir + "/" + chartFilename
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	// Push chart to be deleted.

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push --relative %s %s", chartFilepath, repoName))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

	// Check that pushed chart exists in the bucket.

	obj, err := mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, chartObjectName, obj.Key)

	// Check that pushed chart can be searched, which means it exists in the index.

	cmd, stdout, stderr = command(makeSearchCommand(repoName, chartName))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)

	expected := `test-delete-relative/foo	1.2.3        	1.2.3      	A Helm chart for Kubernetes`
	assert.Contains(t, stdout.String(), expected)

	// Delete chart.

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 delete %s --version %s %s", chartName, chartVersion, repoName))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully deleted the chart from the repository.")

	// Check that chart was actually deleted from the bucket.

	_, err = mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.Equal(t, "NoSuchKey", minio.ToErrorResponse(err).Code)

	// Check that deleted chart cannot be searched, which means it was deleted from the index.

	cmd, stdout, stderr = command(makeSearchCommand(repoName, chartName))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)

	expected = `No results found`
	assert.Contains(t, stdout.String(), expected)
}

func TestDeleteProvenance(t *testing.T) {
	t.Log("Test delete action when chart is pushed with .prov file")

	const (
		repoName        = "test-delete-provenance"
		repoDir         = "charts"
		chartName       = "foo"
		chartVersion    = "1.3.1"
		chartFilename   = chartName + "-" + chartVersion + ".tgz"
		chartFilepath   = "testdata/" + chartFilename
		chartObjectName = repoDir + "/" + chartFilename
		provObjectName  = chartObjectName + ".prov"
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	// Push chart to be deleted.

	cmd := fmt.Sprintf("helm s3 push %s %s", chartFilepath, repoName)
	stdout, stderr, err := runCommand(cmd)
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

	// Check that pushed chart exists in the bucket.

	obj, err := mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, chartObjectName, obj.Key)

	// Check that .prov file exists in the bucket.
	obj, err = mc.StatObject(repoName, provObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, provObjectName, obj.Key)

	// Check that pushed chart can be searched, which means it exists in the index.

	cmd = makeSearchCommand(repoName, chartName)
	stdout, stderr, err = runCommand(cmd)
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)

	expected := `test-delete-provenance/foo	1.3.1        	1.0        	A Helm chart for Kubernetes`
	assert.Contains(t, stdout.String(), expected)

	// Delete chart.

	cmd = fmt.Sprintf("helm s3 delete %s --version %s %s", chartName, chartVersion, repoName)
	stdout, stderr, err = runCommand(cmd)
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully deleted the chart from the repository.")

	// Check that chart was actually deleted from the bucket.

	_, err = mc.StatObject(repoName, chartObjectName, minio.StatObjectOptions{})
	assert.Equal(t, "NoSuchKey", minio.ToErrorResponse(err).Code)

	// Check that .prov file was actually deleted from the bucket.

	_, err = mc.StatObject(repoName, provObjectName, minio.StatObjectOptions{})
	assert.Equal(t, "NoSuchKey", minio.ToErrorResponse(err).Code)

	// Check that deleted chart cannot be searched, which means it was deleted from the index.

	cmd = makeSearchCommand(repoName, chartName)
	stdout, stderr, err = runCommand(cmd)
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)

	expected = `No results found`
	assert.Contains(t, stdout.String(), expected)
}

func TestDeleteMultipleVersions(t *testing.T) {
	t.Log("Test delete with multiple --version values in one run")

	const (
		repoName         = "test-delete-multi-ver"
		repoDir          = "charts"
		chartName        = "foo"
		verA             = "1.2.3"
		verB             = "1.3.1"
		chartFileA       = "foo-" + verA + ".tgz"
		chartFileB       = "foo-" + verB + ".tgz"
		chartFilepathA   = "testdata/" + chartFileA
		chartFilepathB   = "testdata/" + chartFileB
		chartObjectNameA = repoDir + "/" + chartFileA
		chartObjectNameB = repoDir + "/" + chartFileB
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push %s %s", chartFilepathA, repoName))
	require.NoError(t, cmd.Run())
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 push %s %s", chartFilepathB, repoName))
	require.NoError(t, cmd.Run())
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 delete %s --version %s --version %s %s", chartName, verA, verB, repoName))
	require.NoError(t, cmd.Run())
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully deleted 2 chart versions from the repository.")

	_, err := mc.StatObject(repoName, chartObjectNameA, minio.StatObjectOptions{})
	assert.Equal(t, "NoSuchKey", minio.ToErrorResponse(err).Code)
	_, err = mc.StatObject(repoName, chartObjectNameB, minio.StatObjectOptions{})
	assert.Equal(t, "NoSuchKey", minio.ToErrorResponse(err).Code)

	cmd, stdout, stderr = command(makeSearchCommand(repoName, chartName))
	require.NoError(t, cmd.Run())
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "No results found")
}
