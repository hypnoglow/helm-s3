package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/dotaws"
)

const (
	indexTemplate = `apiVersion: v1
entries: {}
generated: {{ .Date }}`
)

func runInit(uri string) {
	tpl := template.New("index")
	tpl, err := tpl.Parse(indexTemplate)
	if err != nil {
		log.Fatalf("failed to parse index.yaml template: %s", err)
	}

	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, map[string]interface{}{"Date": time.Now().Format(time.RFC3339Nano)}); err != nil {
		log.Fatalf("failed to execute index.yaml temlate: %s", err)
	}

	if err := dotaws.ParseCredentials(); err != nil {
		log.Fatalf("failed to parse aws credentials file: %s", err)
	}
	if err := dotaws.ParseConfig(); err != nil {
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

	if _, err := awss3.Upload(ctx, uri+"/index.yaml", buf, awsConfig); err != nil {
		log.Fatalf("failed to upload chart to s3: %s", err)
	}

	fmt.Printf("Initialized empty repository at %s\n", uri)
}
