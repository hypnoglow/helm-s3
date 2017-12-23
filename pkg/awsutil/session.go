package awsutil

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	// awsDisableSSL can be set to true by build tag.
	awsDisableSSL = "false"

	// awsEndpoint can be set to a custom endpoint by build tag.
	awsEndpoint = ""
)

// Config returns AWS config with credentials and parameters taken from
// environment and/or from ~/.aws/* files.
func Session() (*session.Session, error) {

	return session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			DisableSSL:       aws.Bool(awsDisableSSL == "true"),
			S3ForcePathStyle: aws.Bool(true),
			Endpoint:         aws.String(awsEndpoint),
		},
		SharedConfigState: session.SharedConfigEnable,
	})

	// return sess.Config, nil
	// return &aws.Config{
	// 	Credentials: credentials.NewStaticCredentials(
	// 		os.Getenv(envAwsAccessKeyID),
	// 		os.Getenv(envAwsSecretAccessKey),
	// 		os.Getenv(envAwsSessionToken),
	// 	),
	// 	DisableSSL:       aws.Bool(awsDisableSSL == "true"),
	// 	Endpoint:         aws.String(awsEndpoint),
	// 	Region:           aws.String(os.Getenv(envAwsDefaultRegion)),
	// 	S3ForcePathStyle: aws.Bool(true),
	// }, nil
}
