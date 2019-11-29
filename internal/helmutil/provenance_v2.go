package helmutil

import (
	"io"

	"k8s.io/helm/pkg/provenance"
)

func digestV2(in io.Reader) (string, error) {
	return provenance.Digest(in)
}

func digestFileV2(filename string) (string, error) {
	return provenance.DigestFile(filename)
}
