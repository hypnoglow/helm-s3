package helmutil

import (
	"io"
	"os"
)

// Index describes helm chart repo index.
type Index interface {
	// Add adds chart version to the index.
	//
	// Note: this can leave the index in an unsorted state.
	Add(metadata interface{}, filename, baseURL, digest string) error

	// AddOrReplace adds chart version to the index, replacing the version if it exists instead
	// of adding it to the list of versions. Note that helm Add method does not control whether
	// the chart version exists, allowing for duplicates. This methods prevents duplicates
	// by replacing the chart.
	//
	// Note: this can leave the index in an unsorted state.
	AddOrReplace(metadata interface{}, filename, baseURL, digest string) error

	// Delete removes chart version from the index and returns url to the deleted item.
	Delete(name, version string) (url string, err error)

	// Has returns true if the index has an entry for a chart with the given name and exact version.
	Has(name, version string) bool

	// SortEntries sorts the entries by version in descending order.
	SortEntries()

	// MarshalBinary encodes index to a binary form.
	MarshalBinary() (data []byte, err error)

	// UnmarshalBinary decodes index from a binary form.
	UnmarshalBinary(b []byte) error

	// Reader returns io.Reader for index.
	Reader() (io.Reader, error)

	// WriteFile writes an index file to the given destination path.
	WriteFile(dest string, mode os.FileMode) error
}

// NewIndex returns a new Index based either on Helm v2 or Helm v3.
func NewIndex() Index {
	if IsHelm3() {
		return newIndexV3()
	}
	return newIndexV2()
}

// LoadIndex loads index from the file.
func LoadIndex(fpath string) (Index, error) {
	if IsHelm3() {
		return loadIndexV3(fpath)
	}
	return loadIndexV2(fpath)
}
