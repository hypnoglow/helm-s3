package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	t.Log("Test basic init action")

	const (
		repoName        = "test-init"
		repoDir         = "charts"
		indexObjectName = repoDir + "/index.yaml"
		uri             = "s3://test-init/charts"
	)

	setupBucket(t, repoName)
	defer teardownBucket(t, repoName)

	// Run init.

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 init %s", uri))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Initialized empty repository at s3://test-init/charts")

	// Check that initialized repository has index file.

	obj, err := mc.StatObject(repoName, indexObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, indexObjectName, obj.Key)

	// Check that `helm repo add` works.

	cmd, stdout, stderr = command(fmt.Sprintf("helm repo add %s %s", repoName, uri))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), `"test-init" has been added to your repositories`)

	// Check that `helm repo remove` works.

	cmd, stdout, stderr = command(fmt.Sprintf("helm repo remove %s", repoName))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), `"test-init" has been removed from your repositories`)
}

func TestInitForce(t *testing.T) {
	t.Log("Test init action with --force flag")

	const (
		repoName        = "test-init-force"
		repoDir         = "charts"
		indexObjectName = repoDir + "/index.yaml"
		uri             = "s3://" + repoName + "/" + repoDir
	)

	setupBucket(t, repoName)
	defer teardownBucket(t, repoName)

	// Run init first time.

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 init %s", uri))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Initialized empty repository at s3://test-init-force/charts")

	// Check that initialized repository has index file and remember last modification time.

	obj, err := mc.StatObject(repoName, indexObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, indexObjectName, obj.Key)

	lastModified := obj.LastModified

	// Run init again and check that we get error.

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 init %s", uri))
	err = cmd.Run()
	assert.Error(t, err)
	assertEmptyOutput(t, stdout, nil)
	assert.Contains(t, stderr.String(), "The index file already exists under the provided URI")

	// Run init again with --force.

	time.Sleep(time.Second) // To ensure that at least a second passed since last object modification.

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 init %s --force", uri))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Initialized empty repository at s3://test-init-force/charts")

	// Check that index file was overwritten.

	obj, err = mc.StatObject(repoName, indexObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.True(t, obj.LastModified.After(lastModified), "Expected %s to be less than %s", lastModified.String(), obj.LastModified.String())
}

func TestInitIgnoreIfExists(t *testing.T) {
	t.Log("Test init action with --ignore-if-exists flag")

	const (
		repoName        = "test-init-ignore-if-exists"
		repoDir         = "charts"
		indexObjectName = repoDir + "/index.yaml"
		uri             = "s3://" + repoName + "/" + repoDir
	)

	setupBucket(t, repoName)
	defer teardownBucket(t, repoName)

	// Run init first time.

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 init %s", uri))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Initialized empty repository at s3://test-init-ignore-if-exists/charts")

	// Check that initialized repository has index file and remember last modification time.

	obj, err := mc.StatObject(repoName, indexObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, indexObjectName, obj.Key)

	lastModified := obj.LastModified

	// Run init again and check that we get error.

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 init %s", uri))
	err = cmd.Run()
	assert.Error(t, err)
	assertEmptyOutput(t, stdout, nil)
	assert.Contains(t, stderr.String(), "The index file already exists under the provided URI")

	// Run init again with --ignore-if-exists.

	time.Sleep(time.Second) // To ensure that at least a second passed since last object modification.

	cmd, stdout, stderr = command(fmt.Sprintf("helm s3 init %s --ignore-if-exists", uri))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "The index file already exists under the provided URI, ignore init operation.")

	// Check that index file was not overwritten.

	obj, err = mc.StatObject(repoName, indexObjectName, minio.StatObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, lastModified, obj.LastModified)
}

func TestInitForceAndIgnoreIfExists(t *testing.T) {
	t.Log("Test init action with both --force and --ignore-if-exists flags")

	const (
		repoName = "test-init-force-and-ignore-if-exists"
		repoDir  = "charts"
		uri      = "s3://" + repoName + "/" + repoDir
	)

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 init %s --force --ignore-if-exists", uri))
	err := cmd.Run()
	assert.Error(t, err)
	assertEmptyOutput(t, stdout, nil)

	expectedStderr := "The --force and --ignore-if-exists flags are mutually exclusive and cannot be specified together."
	assert.Contains(t, stderr.String(), expectedStderr)
}

func TestInitRepoFileNotFound(t *testing.T) {
	t.Log("Test init action when repo file not found")

	const (
		repoName        = "test-init-repo-file-not-found"
		repoDir         = "charts"
		indexObjectName = repoDir + "/index.yaml"
		uri             = "s3://" + repoName + "/" + repoDir
	)

	setupBucket(t, repoName)
	defer teardownBucket(t, repoName)

	// Run init.

	t.Setenv("HELM_REPOSITORY_CONFIG", "testdata/file-not-exists.yaml")

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 init %s", uri))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Initialized empty repository at "+uri)

	// Skip other checks because they are already covered by TestInit.
}
