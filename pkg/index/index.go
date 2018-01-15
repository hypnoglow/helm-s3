package index

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ghodss/yaml"

	"k8s.io/helm/pkg/repo"
)

// Index of a helm chart repository.
type Index struct {
	*repo.IndexFile
}

// Reader returns io.Reader for index.
func (idx *Index) Reader() (io.Reader, error) {
	b, err := idx.Bytes()
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

// Bytes returns index in bytes representation.
func (idx *Index) Bytes() ([]byte, error) {
	return yaml.Marshal(idx)
}

// Delete removes chart version from index and returns deleted item.
func (idx *Index) Delete(name, version string) (*repo.ChartVersion, error) {
	for chartName, chartVersions := range idx.Entries {
		if chartName != name {
			continue
		}

		for i, chartVersion := range chartVersions {
			if chartVersion.Version == version {
				idx.Entries[chartName] = append(
					idx.Entries[chartName][:i],
					idx.Entries[chartName][i+1:]...,
				)
				return chartVersion, nil
			}
		}
	}

	return nil, fmt.Errorf("chart %s version %s not found in index", name, version)
}

// New returns a new index.
func New() *Index {
	return &Index{
		repo.NewIndexFile(),
	}
}

// LoadBytes returns an index read from bytes.
func LoadBytes(b []byte) (*Index, error) {
	i := &repo.IndexFile{}
	if err := yaml.Unmarshal(b, i); err != nil {
		return nil, err
	}
	i.SortEntries()
	return &Index{i}, nil
}
