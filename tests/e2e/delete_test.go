package e2e

import (
	"fmt"
	"testing"

	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/assert"
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
