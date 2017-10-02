package main

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/dotaws"
)

const (
	tmpCurrentIndex = "/tmp/current-index.yaml"
)

func runPush(chartPath string, repoName string) {
	fpath, err := filepath.Abs(chartPath)
	if err != nil {
		log.Fatalf("failed to locate chart archive: %s", err)
	}

	dir := filepath.Dir(fpath)
	chartFilename := filepath.Base(fpath)

	os.Chdir(dir)

	chartFile, err := os.Open(chartFilename)
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	// TODO: check if it is a real chart.

	repoURL, err := lookupRepoURI(repoName)
	if err != nil {
		log.Fatalf("failed to discover repo URL: %s", err)
	}

	if err = dotaws.ParseCredentials(); err != nil {
		log.Fatalf("failed to parse aws credentials file: %s", err)
	}
	if err = dotaws.ParseConfig(); err != nil {
		log.Fatalf("failed to parse aws config file: %s", err)
	}
	awsConfig := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			os.Getenv(envAwsAccessKeyID),
			os.Getenv(envAwsSecretAccessKey),
			"",
		),
		Region: aws.String(os.Getenv(envAWsDefaultRegion)),
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	b, err := awss3.FetchRaw(ctx, repoURL+"/index.yaml", awsConfig)
	if err != nil {
		log.Fatalf("failed to fetch current repo index: %s", err)
	}
	if err := ioutil.WriteFile(tmpCurrentIndex, b, 0755); err != nil {
		log.Fatalf("faield to write current repo index to %s: %s", tmpCurrentIndex, err)
	}

	cmd := exec.Command("helm", "repo", "index", "--url", repoURL, "--merge", tmpCurrentIndex, ".")
	if err = cmd.Run(); err != nil {
		log.Fatalf("failed to index chart: %s", err)
	}
	indexFile, err := os.Open("./index.yaml")
	if err != nil {
		log.Fatalf("failed to open new index.yaml file: %s", err)
	}

	// Finally, upload both chart file and index.

	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout*2)
	defer cancel()
	if _, err := awss3.Upload(ctx, repoURL+"/"+chartFilename, chartFile, awsConfig); err != nil {
		log.Fatalf("failed to upload chart to s3: %s", err)
	}
	if _, err := awss3.Upload(ctx, repoURL+"/index.yaml", indexFile, awsConfig); err != nil {
		log.Fatalf("failed to upload index to s3: %s", err)
	}
}

func lookupRepoURI(name string) (string, error) {
	cmd := exec.Command("helm", "repo", "list")
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, `failed to exec "helm repoURL list": %s`)
	}

	var repoURL string
	buf := bytes.NewBuffer(out)
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		str := scanner.Text()
		if strings.HasPrefix(str, name) {
			repoURL = strings.TrimSpace(strings.TrimPrefix(str, name))
		}
	}
	if err := scanner.Err(); err != nil {
		return "", errors.Wrap(err, "failed to scan helm repo list")
	}
	if repoURL == "" {
		return "", errors.Errorf("repoURL with name %s not found, try `helm repoURL add %s <uri>`", name, name)
	}

	return repoURL, nil
}
