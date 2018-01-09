package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
	"github.com/hypnoglow/helm-s3/pkg/index"
)

func runInit(uri string) error {
	r, err := index.New().Reader()
	if err != nil {
		return errors.WithMessage(err, "get index reader")
	}

	storage := awss3.New()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if err := storage.PutIndex(ctx, uri, r); err != nil {
		return errors.WithMessage(err, "upload index to s3")
	}

	return nil
}
