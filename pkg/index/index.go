package index

import (
	"bytes"
	"io"

	"github.com/ghodss/yaml"

	"k8s.io/helm/pkg/repo"
)

// Index of a helm chart repository.
type Index struct {
	*repo.IndexFile
}

// Reader returns io.Reader for index.
func (i Index) Reader() (io.Reader, error) {
	b, err := yaml.Marshal(i)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

// New returns a new helm chart repository index.
func New() Index {
	return Index{
		repo.NewIndexFile(),
	}
}
