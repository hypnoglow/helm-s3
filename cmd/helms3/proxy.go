package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/dotaws"
)

func runProxy(uri string) {
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

	b, err := awss3.FetchRaw(ctx, uri, awsConfig)
	if err != nil {
		log.Fatalf("failed to fetch from s3: %s", err)
	}

	fmt.Print(string(b))
}
