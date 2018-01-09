package main

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"

	"github.com/hypnoglow/helm-s3/pkg/awss3"
)

func runProxy(uri string) error {

	storage := awss3.New()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	b, err := storage.FetchRaw(ctx, uri)
	if err != nil {
		if strings.HasSuffix(uri, "index.yaml") && err == awss3.ErrObjectNotFound {
			return fmt.Errorf("The index file does not exist by the path %s. If you haven't initialized the repository yet, try running \"helm s3 init %s\"", uri, path.Dir(uri))
		}
		return errors.WithMessage(err, "fetch from s3")
	}

	fmt.Print(string(b))
	return nil
}
