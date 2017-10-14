package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/awsutil"
)

func runProxy(uri string) error {
	awsConfig, err := awsutil.Config()
	if err != nil {
		return errors.WithMessage(err, "get aws config")
	}

	storage := awss3.NewStorage(awsConfig)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	b, err := storage.FetchRaw(ctx, uri)
	if err != nil {
		return errors.WithMessage(err, "fetch from s3")
	}

	fmt.Print(string(b))
	return nil
}
