package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-ini/ini"
	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/pkg/s3object"
)

const (
	envAwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	envAwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	envAWsDefaultRegion   = "AWS_DEFAULT_REGION"
)

var (
	credFile = os.ExpandEnv("$HOME/.aws/credentials")
	confFile = os.ExpandEnv("$HOME/.aws/config")
)

func main() {
	if len(os.Args) < 5 {
		fmt.Print("The direct use of \"helm s3\" command is currently not supported.")
		return
	}

	parseCreds()
	parseConfig()

	uri := os.Args[4]
	result, err := get(uri)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(os.Stderr, string(result))
	fmt.Print(string(result))
}

func get(uri string) ([]byte, error) {
	awsConfig := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			os.Getenv(envAwsAccessKeyID),
			os.Getenv(envAwsSecretAccessKey),
			"",
		),
		Region: aws.String(os.Getenv(envAWsDefaultRegion)),
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new aws session")
	}

	bucket, key := s3object.Parse(uri)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	buf := &aws.WriteAtBuffer{}
	_, err = s3manager.
		NewDownloader(sess).
		DownloadWithContext(ctx, buf, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		return nil, errors.Wrap(err, "failed to upload file to s3")
	}

	return buf.Bytes(), nil
}

func parseCreds() {
	f, err := os.Open(credFile)
	if err != nil {
		log.Print("[DEBUG] failed to read aws credentials file:", err)
		return
	}

	fmt.Printf("DEBUG: %#v\n", "blya")

	il, err := ini.Load(f)
	if err != nil {
		log.Fatal(err)
	}

	sec, err := il.GetSection("default")
	if err != nil {
		log.Fatal(err)
	}

	accessKeyID, err := sec.GetKey("aws_access_key_id")
	if err != nil {
		log.Fatal(err)
	}

	os.Setenv(envAwsAccessKeyID, accessKeyID.String())

	secretAccessKey, err := sec.GetKey("aws_secret_access_key")
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv(envAwsSecretAccessKey, secretAccessKey.String())
}

func parseConfig() {
	f, err := os.Open(confFile)
	if err != nil {
		log.Print("[DEBUG] failed to read aws config file:", err)
		return
	}

	il, err := ini.Load(f)
	if err != nil {
		log.Fatal(err)
	}

	sec, err := il.GetSection("default")
	if err != nil {
		log.Fatal(err)
	}

	region, err := sec.GetKey("region")
	if err != nil {
		log.Fatal(err)
	}

	os.Setenv(envAWsDefaultRegion, region.String())
}
