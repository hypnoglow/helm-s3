package dotaws

import (
	"os"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
)

const (
	credentialsFile = "$HOME/.aws/credentials"

	envAwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	envAwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
)

func ParseCredentials(profile string) error {
	f, err := os.Open(os.ExpandEnv(credentialsFile))
	if err != nil {
		if err == os.ErrNotExist {
			return nil
		}
		return errors.Wrap(err, "failed to open aws credentials file")
	}

	il, err := ini.Load(f)
	if err != nil {
		return errors.Wrapf(err, "failed to load file %s as ini", credentialsFile)
	}

	sectionName := "default"
	if profile != "" {
		sectionName = profile
	}

	sec, err := il.GetSection(sectionName)
	if err != nil {
		return errors.Wrap(err, `aws credentials file has no "default" section`)
	}

	accessKeyID, err := sec.GetKey("aws_access_key_id")
	if err != nil {
		return errors.Wrap(err, `aws credentials file "default" section has no key "aws_access_key_id"`)
	}

	secretAccessKey, err := sec.GetKey("aws_secret_access_key")
	if err != nil {
		return errors.Wrap(err, `aws credentials file "default" section has no key "aws_secret_access_key"`)
	}

	os.Setenv(envAwsAccessKeyID, accessKeyID.String())
	os.Setenv(envAwsSecretAccessKey, secretAccessKey.String())
	return nil
}
