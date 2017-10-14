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

// New returns a new index.
func New() Index {
	return Index{
		repo.NewIndexFile(),
	}
}

// LoadBytes returns an index read from bytes.
func LoadBytes(b []byte) (Index, error) {
	i := &repo.IndexFile{}
	if err := yaml.Unmarshal(b, i); err != nil {
		return Index{}, err
	}
	i.SortEntries()
	return Index{i}, nil
}
