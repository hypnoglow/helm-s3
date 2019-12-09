package helmutil

import (
	"io"

	"helm.sh/helm/v3/pkg/provenance"
)

func digestV3(in io.Reader) (string, error) {
	return provenance.Digest(in)
}

func digestFileV3(filename string) (string, error) {
	return provenance.DigestFile(filename)
}
