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
	envAwsSessionToken    = "AWS_SESSION_TOKEN"
	envAwsDefaultRegion   = "AWS_DEFAULT_REGION"

	envAwsProfile = "AWS_PROFILE"
)

var (
	// awsDisableSSL can be set to true by build tag.
	awsDisableSSL = "false"

	// awsEndpoint can be set to a custom endpoint by build tag.
	awsEndpoint = ""

	// awsSessionToken retrieved from http://169.254.169.254/latest/meta-data/iam/security-credentials/ ( IAM Instance Profile )
	awsSessionToken = ""
)

// Config returns AWS config with credentials and parameters taken from
// environment and/or from ~/.aws/* files.
func Config() (*aws.Config, error) {
	profile := os.Getenv(envAwsProfile)

	if os.Getenv(envAwsAccessKeyID) == "" && os.Getenv(envAwsSecretAccessKey) == "" {
		if err := dotaws.ParseCredentials(profile); err != nil {
			return nil, errors.Wrap(err, "failed to parse aws credentials file")
		}
	}

	if os.Getenv(envAwsDefaultRegion) == "" {
		if err := dotaws.ParseConfig(profile); err != nil {
			return nil, errors.Wrap(err, "failed to parse aws config file")
		}
	}

	if os.Getenv(envAwsSessionToken) != "" {
		awsSessionToken = os.Getenv(envAwsSessionToken)
	}

	return &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			os.Getenv(envAwsAccessKeyID),
			os.Getenv(envAwsSecretAccessKey),
			awsSessionToken,
		),
		DisableSSL:       aws.Bool(awsDisableSSL == "true"),
		Endpoint:         aws.String(awsEndpoint),
		Region:           aws.String(os.Getenv(envAwsDefaultRegion)),
		S3ForcePathStyle: aws.Bool(true),
	}, nil
}
