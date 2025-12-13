package awsutil

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	// awsEndpoint can be set to a custom endpoint to use alternative AWS S3
	// server like minio (https://minio.io).
	awsEndpoint = "AWS_ENDPOINT"

	// awsDisableSSL can be set to true to disable SSL for AWS S3 server.
	awsDisableSSL = "AWS_DISABLE_SSL"

	// awsBucketLocation can be set to an AWS region to force the session region
	// if AWS_DEFAULT_REGION and AWS_REGION cannot be trusted.
	awsBucketLocation = "HELM_S3_REGION"
)

// SessionOption is an option for session.
type SessionOption func(*config.LoadOptions) error

// AssumeRoleTokenProvider is an option for setting custom assume role token provider.
func AssumeRoleTokenProvider(provider func() (string, error)) SessionOption {
	return func(options *config.LoadOptions) error {
		// Note: In AWS SDK v2, assume role token provider configuration
		// is more complex and should be set during actual credential usage
		// For now, we'll just return nil and handle this in credential flow
		return nil
	}
}

// DynamicBucketRegion is an option for determining the Helm S3 bucket's AWS
// region dynamically thus allowing the mixed use of buckets residing in
// different regions without requiring manual updates on the HELM_S3_REGION,
// AWS_REGION, or AWS_DEFAULT_REGION environment variables.
//
// This HEAD bucket solution works with all kinds of S3 URIs containing
// the bucket name in the host part.
//
// The basic idea behind the HEAD bucket solution and the "official
// confirmation" this behavior is expected and supported came from a comment on
// the AWS SDK Go repository:
// https://github.com/aws/aws-sdk-go/issues/720#issuecomment-243891223
func DynamicBucketRegion(s3URL string) SessionOption {
	return func(options *config.LoadOptions) error {
		parsedS3URL, err := url.Parse(s3URL)
		if err != nil || parsedS3URL.Host == "" {
			return nil
		}

		// Note: The dummy credentials are required in case no other credential
		// provider is found, but even if the HEAD bucket request fails and
		// returns a non-200 status code indicating no access to the bucket, the
		// actual bucket region is returned in a response header.
		//
		// Note: A signing region **MUST** be configured, otherwise the signed
		// request fails. The configured region itself is irrelevant, the
		// endpoint officially works and returns the bucket region in a response
		// header regardless of whether the signing region matches the bucket's
		// region.
		//
		// Note: The default S3 endpoint **MUST** be configured to avoid making
		// the request region specific thus avoiding regional redirect responses
		// (301 Permanently moved) on HEAD bucket. This setting is only required
		// because any other region than "us-east-1" would configure a
		// region-specific endpoint as well, so it's more safe to explicitly
		// configure the default endpoint.
		//
		// Source:
		// https://github.com/aws/aws-sdk-go/issues/720#issuecomment-243891223
		ctx := context.Background()
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion("us-east-1"),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:           "https://s3.amazonaws.com",
						SigningRegion: "us-east-1",
					}, nil
				},
			)),
		)
		if err != nil {
			return nil
		}

		s3Client := s3.NewFromConfig(cfg)

		bucketRegionHeader := "X-Amz-Bucket-Region"
		input := &s3.HeadBucketInput{
			Bucket: aws.String(parsedS3URL.Host),
		}

		resp, err := s3Client.HeadBucket(ctx, input)
		if err != nil || resp == nil {
			return nil
		}

		// Extract the region from the response metadata
		if httpResp, ok := resp.ResultMetadata.Get("http.Response").(*http.Response); ok {
			if region := httpResp.Header.Get(bucketRegionHeader); region != "" {
				options.Region = region
			}
		}

		return nil
	}
}

// Session returns an AWS config as described in AWS SDK for Go v2 documentation
func Session(opts ...SessionOption) (aws.Config, error) {
	ctx := context.Background()

	// Build the config options
	configOpts := []func(*config.LoadOptions) error{
		config.WithSharedConfigProfile(""),
	}

	// Add assume role token provider
	configOpts = append(configOpts, AssumeRoleTokenProvider(StderrTokenProvider))

	// Set region if specified
	bucketRegion := os.Getenv(awsBucketLocation)
	if bucketRegion != "" {
		configOpts = append(configOpts, config.WithRegion(bucketRegion))
	}

	// Add custom options
	for _, opt := range opts {
		configOpts = append(configOpts, opt)
	}

	// Load the configuration
	cfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return aws.Config{}, err
	}

	// Configure endpoint and SSL settings
	endpoint := os.Getenv(awsEndpoint)
	disableSSL := os.Getenv(awsDisableSSL) == "true"

	if endpoint != "" || disableSSL {
		cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				if endpoint != "" {
					return aws.Endpoint{
						URL:               endpoint,
						HostnameImmutable: true,
						Source:            aws.EndpointSourceCustom,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			},
		)
	}

	// Set HTTP client with SSL configuration
	if disableSSL {
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		cfg.HTTPClient = httpClient
	}

	return cfg, nil
}
