package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/minio/minio-go/v6"
)

var mc *minio.Client

func TestMain(m *testing.M) {
	setup()
	defer teardown()

	os.Exit(m.Run())
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

func setupRepo(t *testing.T, name, dir string) {
	url := fmt.Sprintf("s3://%s", name)
	if dir != "" {
		url += "/" + dir
	}

	if exists, err := mc.BucketExists(name); err != nil {
		t.Fatalf("check bucket exists: %v", err.Error())
	} else if !exists {
		if err = mc.MakeBucket(name, ""); err != nil {
			t.Fatalf("create bucket: %v", err)
		}
	}

	if out, err := exec.Command("helm", "s3", "init", url).CombinedOutput(); err != nil {
		t.Fatalf("init repo: %v %q", err, string(out))
	}

	if out, err := exec.Command("helm", "repo", "add", name, url).CombinedOutput(); err != nil {
		t.Fatalf("add repo: %v %q", err, string(out))
	}
}

func teardownRepo(t *testing.T, name string) {
	if err := exec.Command("helm", "repo", "remove", name).Run(); err != nil {
		t.Fatalf("remove repo: %v", err)
	}
}

func command(c string) (cmd *exec.Cmd, stdout, stderr *bytes.Buffer) {
	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}
	args := strings.Split(c, " ")

	cmd = exec.Command(args[0], args[1:]...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return
}
