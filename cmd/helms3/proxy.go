package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/awsutil"
)

func runProxy(uri string) {
	awsConfig, err := awsutil.Config()
	if err != nil {
		log.Fatalf("failed to get aws config: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	b, err := awss3.FetchRaw(ctx, uri, awsConfig)
	if err != nil {
		log.Fatalf("failed to fetch from s3: %s", err)
	}

	fmt.Print(string(b))
}
