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
	envAwsSessionToken    = "AWS_SESSION_TOKEN"
)

func ParseCredentials(profile string) error {
	f, err := os.Open(os.ExpandEnv(credentialsFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(err, "open aws credentials file")
	}

	il, err := ini.Load(f)
	if err != nil {
		return errors.Wrapf(err, "load file %s as ini", credentialsFile)
	}

	sectionName := "default"
	if profile != "" {
		sectionName = profile
	}

	sec, err := il.GetSection(sectionName)
	if err != nil {
		return errors.Wrapf(err, "aws credentials file has no section %q", sectionName)
	}

	if accessKeyID, err := sec.GetKey("aws_access_key_id"); err == nil {
		os.Setenv(envAwsAccessKeyID, accessKeyID.String())
	}

	if secretAccessKey, err := sec.GetKey("aws_secret_access_key"); err == nil {
		os.Setenv(envAwsSecretAccessKey, secretAccessKey.String())
	}

	if awsSessionToken, err := sec.GetKey("aws_session_token"); err == nil {
		os.Setenv(envAwsSessionToken, awsSessionToken.String())
	}

	return nil
}
