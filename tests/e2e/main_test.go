package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hypnoglow/helm-s3/internal/helmutil"
)

var mc *minio.Client

func TestMain(m *testing.M) {
	setup()

	code := m.Run()

	teardown()

	os.Exit(code)
}

func setup() {
	if os.Getenv("AWS_ENDPOINT") == "" {
		panic("AWS_ENDPOINT is empty")
	}

	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		panic("AWS_ACCESS_KEY_ID is empty")
	}

	if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		panic("AWS_SECRET_ACCESS_KEY is empty")
	}

	var err error
	mc, err = minio.New(
		os.Getenv("AWS_ENDPOINT"),
		os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
		false,
	)
	if err != nil {
		panic("create minio client: " + err.Error())
	}
}

func teardown() {

}

// helper functions

func setupBucket(t *testing.T, name string) {
	t.Helper()

	exists, err := mc.BucketExists(name)
	require.NoError(t, err, "check if bucket exists")
	if exists {
		teardownBucket(t, name)
	}

	err = mc.MakeBucket(name, "")
	require.NoError(t, err, "create bucket")
}

func teardownBucket(t *testing.T, name string) {
	t.Helper()

	done := make(chan struct{})
	defer close(done)

	for obj := range mc.ListObjectsV2(name, "", true, done) {
		err := mc.RemoveObject(name, obj.Key)
		assert.NoError(t, err)
	}

	err := mc.RemoveBucket(name)
	require.NoError(t, err, "remove bucket")
}

func setupRepo(t *testing.T, name, dir string) { //nolint:unparam
	t.Helper()

	setupBucket(t, name)

	url := fmt.Sprintf("s3://%s", name)
	if dir != "" {
		url += "/" + dir
	}

	out, err := exec.Command("helm", "s3", "init", url).CombinedOutput()
	require.NoError(t, err, "helm s3 init: %s", string(out))

	out, err = exec.Command("helm", "repo", "add", name, url).CombinedOutput()
	require.NoError(t, err, "helm repo add: %s", string(out))
}

func teardownRepo(t *testing.T, name string) {
	t.Helper()

	err := exec.Command("helm", "repo", "remove", name).Run()
	require.NoError(t, err)

	teardownBucket(t, name)
}

func command(c string) (cmd *exec.Cmd, stdout, stderr *bytes.Buffer) {
	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}
	args := strings.Split(c, " ")

	cmd = exec.Command(args[0], args[1:]...) //nolint:gosec
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return
}

// For helm v2, the command is `helm search foo/bar`.
// For helm v3, the command is `helm search repo foo/bar`.
func makeSearchCommand(repoName, chartName string) string {
	c := "helm search"

	helmutil.SetupHelm()
	if helmutil.IsHelm3() {
		c += " repo"
	}

	return fmt.Sprintf("%s %s/%s", c, repoName, chartName)
}
