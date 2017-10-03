package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/awsutil"
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

	awsConfig, err := awsutil.Config()
	if err != nil {
		log.Fatalf("failed to get aws config: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if _, err := awss3.Upload(ctx, uri+"/index.yaml", buf, awsConfig); err != nil {
		log.Fatalf("failed to upload chart to s3: %s", err)
	}

	fmt.Printf("Initialized empty repository at %s\n", uri)
}
