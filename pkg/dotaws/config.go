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
		if err == os.ErrNotExist {
			return nil
		}
		return errors.Wrap(err, "failed to open aws config file")
	}

	il, err := ini.Load(f)
	if err != nil {
		return errors.Wrapf(err, "failed to load file %s as ini", configFile)
	}

	sectionName := "default"
	if profile != "" {
		sectionName = fmt.Sprintf("profile %s", profile)
	}

	sec, err := il.GetSection(sectionName)
	if err != nil {
		return errors.Wrap(err, `aws config file has no "default" section`)
	}

	region, err := sec.GetKey("region")
	if err != nil {
		return errors.Wrap(err, `aws config file '"default" section has no key "region"`)
	}

	os.Setenv(envAWsDefaultRegion, region.String())
	return nil
}
