package helmutil

import (
	"io"
)

// Chart describes a helm chart.
type Chart interface {
	// Name returns chart name.
	// Example: "foo".
	Name() string

	// Version returns chart version.
	// Example: "0.1.0".
	Version() string

	// Metadata returns chart metadata.
	Metadata() ChartMetadata
}

// LoadChart returns chart loaded from the file system by path.
func LoadChart(fpath string) (Chart, error) {
	if IsHelm3() {
		return loadChartV3(fpath)
	}
	return loadChartV2(fpath)
}

// LoadArchive returns chart loaded from the archive file reader.
func LoadArchive(r io.Reader) (Chart, error) {
	if IsHelm3() {
		return loadArchiveV3(r)
	}
	return loadArchiveV2(r)
}

// ChartMetadata describes helm chart metadata.
type ChartMetadata interface {
	// MarshalJSON marshals chart metadata to JSON.
	MarshalJSON() ([]byte, error)

	// UnmarshalJSON unmarshals chart metadata from JSON.
	UnmarshalJSON([]byte) error

	// Value returns underlying chart metadata value.
	Value() interface{}
}

// NewChartMetadata returns a new helm chart metadata.
func NewChartMetadata() ChartMetadata {
	if IsHelm3() {
		return newChartMetadataV3()
	}
	return newChartMetadataV2()
}
