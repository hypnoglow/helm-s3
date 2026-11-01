package e2e

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hypnoglow/helm-s3/internal/helmutil"
)

func TestHelmDependencyUpdate(t *testing.T) {
	t.Log("Test helm dependency update")

	helmutil.SetupHelm()
	if !helmutil.IsHelm3() {
		t.Log("This test only supports Helm v3, skipping...")
		return
	}

	const (
		repoName = "acme-corp"
		repoDir  = "charts"

		fooChartFilename = "foo-1.2.3.tgz"
		fooChartFilepath = "testdata/" + fooChartFilename

		barChartDirname = "bar"
		barChartDirpath = "testdata/" + barChartDirname
	)

	// Set up

	setupRepo(t, repoName, repoDir)
	defer teardownRepo(t, repoName)

	// Test scenario

	// 1. Push 'foo' chart to the repo.

	cmd, stdout, stderr := command(fmt.Sprintf("helm s3 push %s %s", fooChartFilepath, repoName))
	err := cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)
	assert.Contains(t, stdout.String(), "Successfully uploaded the chart to the repository.")

	// 2. Ensure that 'bar' chart has empty 'charts' directory.

	assert.NoError(t, os.RemoveAll(barChartDirpath+"/charts"))
	assert.NoError(t, os.Mkdir(barChartDirpath+"/charts", 0o750))

	// 3. Run `helm dep up`.

	cmd, stdout, stderr = command(fmt.Sprintf("helm dependency update --skip-refresh %s", barChartDirpath))
	err = cmd.Run()
	assert.NoError(t, err)
	assertEmptyOutput(t, nil, stderr)

	expected := "Saving 1 charts\n" +
		"Downloading foo from repo s3://acme-corp/charts\n" +
		"Deleting outdated charts\n"
	assert.Contains(t, stdout.String(), expected)

	// 4. Check that 'charts' directory is populated in 'bar' chart.

	fi, err := os.Stat(barChartDirpath + "/charts/foo-1.2.3.tgz")
	assert.NoError(t, err)
	assert.Equal(t, "foo-1.2.3.tgz", fi.Name())
	assert.Equal(t, false, fi.IsDir())
}
