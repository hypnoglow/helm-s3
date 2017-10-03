package awsutil

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/pkg/dotaws"
)

const (
	envAwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	envAwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	envAWsDefaultRegion   = "AWS_DEFAULT_REGION"
)

// Config returns AWS config with credentials and parameters taken from
// environment and/or from ~/.aws/* files.
func Config() (*aws.Config, error) {
	if os.Getenv(envAwsAccessKeyID) == "" && os.Getenv(envAwsSecretAccessKey) == "" {
		if err := dotaws.ParseCredentials(); err != nil {
			return nil, errors.Wrap(err, "failed to parse aws credentials file")
		}
	}

	if os.Getenv(envAWsDefaultRegion) == "" {
		if err := dotaws.ParseConfig(); err != nil {
			return nil, errors.Wrap(err, "failed to parse aws config file")
		}
	}

	return &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			os.Getenv(envAwsAccessKeyID),
			os.Getenv(envAwsSecretAccessKey),
			"",
		),
		Region: aws.String(os.Getenv(envAWsDefaultRegion)),
	}, nil
}
