package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/internal/awss3"
	"github.com/hypnoglow/helm-s3/internal/awsutil"
)

type proxyCmd struct {
	uri string
}

const indexYaml = "index.yaml"

func (act proxyCmd) Run(ctx context.Context) error {
	sess, err := awsutil.Session(awsutil.AssumeRoleTokenProvider(awsutil.StderrTokenProvider))
	if err != nil {
		return err
	}
	storage := awss3.New(sess)

	b, err := storage.FetchRaw(ctx, act.uri)
	if err != nil {
		if strings.HasSuffix(act.uri, indexYaml) && err == awss3.ErrObjectNotFound {
			return fmt.Errorf("The index file does not exist by the path %s. If you haven't initialized the repository yet, try running \"helm s3 init %s\"", act.uri, act.uri[0:len(indexYaml)])
		}
		return errors.WithMessage(err, "fetch from s3")
	}

	fmt.Print(string(b))
	return nil
}
