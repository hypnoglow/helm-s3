package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/awsutil"
	"github.com/hypnoglow/helm-s3/pkg/index"
)

func runInit(uri string) error {
	r, err := index.New().Reader()
	if err != nil {
		return errors.WithMessage(err, "get index reader")
	}

	awsConfig, err := awsutil.Config()
	if err != nil {
		return errors.WithMessage(err, "get aws config")
	}

	storage := awss3.NewStorage(awsConfig)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if _, err := storage.Upload(ctx, uri+"/index.yaml", r); err != nil {
		return errors.WithMessage(err, "upload index to s3")
	}

	return nil

}
