package dotaws

import (
	"fmt"
	"os"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
)

const (
	configFile = "$HOME/.aws/config"

	envAWsDefaultRegion = "AWS_DEFAULT_REGION"
)

func ParseConfig(profile string) error {
	f, err := os.Open(os.ExpandEnv(configFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(err, "open aws config file")
	}

	il, err := ini.Load(f)
	if err != nil {
		return errors.Wrapf(err, "load file %s as ini", configFile)
	}

	sectionName := "default"
	if profile != "" {
		sectionName = fmt.Sprintf("profile %s", profile)
	}

	sec, err := il.GetSection(sectionName)
	if err != nil {
		return errors.Wrapf(err, "aws config file has no section %q", sectionName)
	}

	if region, err := sec.GetKey("region"); err == nil {
		os.Setenv(envAWsDefaultRegion, region.String())
	}

	return nil
}
