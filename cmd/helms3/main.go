package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/dotaws"
)

const (
	envAwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	envAwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	envAWsDefaultRegion   = "AWS_DEFAULT_REGION"

	defaultTimeout = time.Second * 5
)

func main() {
	if len(os.Args) < 5 {
		fmt.Print("The direct use of \"helm s3\" command is currently not supported.")
		return
	}

	if err := dotaws.ParseCredentials(); err != nil {
		log.Fatalf("failed to parse aws credentials file: %s", err)
	}
	if err := dotaws.ParseConfig(); err != nil {
		log.Fatalf("failed to parse aws config file: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	uri := os.Args[4]
	awsConfig := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			os.Getenv(envAwsAccessKeyID),
			os.Getenv(envAwsSecretAccessKey),
			"",
		),
		Region: aws.String(os.Getenv(envAWsDefaultRegion)),
	}

	b, err := awss3.FetchRaw(ctx, uri, awsConfig)
	if err != nil {
		log.Fatalf("failed to fetch from s3: %s", err)
	}

	fmt.Print(string(b))
}
