package awsutil

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	// awsEndpoint can be set to a custom endpoint to use alternative AWS S3
	// server like minio (https://minio.io).
	awsEndpoint = "AWS_ENDPOINT"

	// awsDisableSSL can be set to true to disable SSL for AWS S3 server.
	awsDisableSSL = "AWS_DISABLE_SSL"

	// awsSSO can be set to true to enable the SSO credential provider
	awsSSO = "AWS_SSO"
)

// SessionOption is an option for session.
type SessionOption func(*session.Options)

// AssumeRoleTokenProvider is an option for setting custom assume role token provider.
func AssumeRoleTokenProvider(provider func() (string, error)) SessionOption {
	return func(options *session.Options) {
		options.AssumeRoleTokenProvider = provider
	}
}

// Session returns an AWS session as described http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html
func Session(opts ...SessionOption) (*session.Session, error) {
	disableSSL := false
	if os.Getenv(awsDisableSSL) == "true" {
		disableSSL = true
	}

	so := session.Options{
		Config: aws.Config{
			DisableSSL:       aws.Bool(disableSSL),
			S3ForcePathStyle: aws.Bool(true),
			Endpoint:         aws.String(os.Getenv(awsEndpoint)),
		},
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: StderrTokenProvider,
	}

	if os.Getenv(awsSSO) == "true" {
		ssoCredentialProvider, err := NewSSOCredentialProvider(os.Getenv("AWS_PROFILE"))
		if err != nil {
			return nil, err
		}
		so.Config.Credentials = credentials.NewCredentials(ssoCredentialProvider)
	}

	for _, opt := range opts {
		opt(&so)
	}

	return session.NewSessionWithOptions(so)
}
