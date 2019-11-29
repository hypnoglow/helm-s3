package helmutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/helm/pkg/urlutil" // Note that this is from Helm v2 SDK because in Helm v3 this package is internal.
	"sigs.k8s.io/yaml"
)

type IndexV3 struct {
	index *repo.IndexFile
}

func (idx *IndexV3) Add(metadata interface{}, filename, baseURL, digest string) error {
	md, ok := metadata.(*chart.Metadata)
	if !ok {
		return errors.New("metadata is not *chart.Metadata")
	}

	idx.index.Add(md, filename, baseURL, digest)
	return nil
}

func (idx *IndexV3) AddOrReplace(metadata interface{}, filename, baseURL, digest string) error {
	// TODO: this looks like a workaround.
	// Think how we can rework this in the future.
	// Ref: https://github.com/kubernetes/helm/issues/3230

	// TODO: this code is the same as for Helm v2, only chart.Medata struct is from Helm v3 SDK.
	// We probably should reduce duplicate code .

	md, ok := metadata.(*chart.Metadata)
	if !ok {
		return errors.New("metadata is not *chart.Metadata")
	}

	u := filename
	if baseURL != "" {
		var err error
		_, file := filepath.Split(filename)
		u, err = urlutil.URLJoin(baseURL, file)
		if err != nil {
			u = filepath.Join(baseURL, file)
		}
	}
	cr := &repo.ChartVersion{
		URLs:     []string{u},
		Metadata: md,
		Digest:   digest,
		Created:  time.Now(),
	}

	// If no chart with such name exists in the index, just create a new
	// list of versions.
	entry, ok := idx.index.Entries[md.Name]
	if !ok {
		idx.index.Entries[md.Name] = repo.ChartVersions{cr}
		return nil
	}

	chartSemVer, err := semver.NewVersion(md.Version)
	if err != nil {
		return err
	}

	// If such version exists, replace it.
	for i, v := range entry {
		itemSemVer, err := semver.NewVersion(v.Version)
		if err != nil {
			return err
		}

		if chartSemVer.Equal(itemSemVer) {
			idx.index.Entries[md.Name][i] = cr
			return nil
		}
	}

	// Otherwise just add to the list of versions
	idx.index.Entries[md.Name] = append(entry, cr)
	return nil
}

func (idx *IndexV3) Delete(name, version string) (url string, err error) {
	for chartName, chartVersions := range idx.index.Entries {
		if chartName != name {
			continue
		}

		for i, chartVersion := range chartVersions {
			if chartVersion.Version == version {
				idx.index.Entries[chartName] = append(
					idx.index.Entries[chartName][:i],
					idx.index.Entries[chartName][i+1:]...,
				)
				if len(chartVersion.URLs) > 0 {
					return chartVersion.URLs[0], nil
				}
				return "", nil
			}
		}
	}

	return "", fmt.Errorf("chart %s version %s not found in index", name, version)
}

func (idx *IndexV3) Has(name, version string) bool {
	return idx.index.Has(name, version)
}

func (idx *IndexV3) SortEntries() {
	idx.index.SortEntries()
}

func (idx *IndexV3) MarshalBinary() (data []byte, err error) {
	return yaml.Marshal(idx.index)
}

func (idx *IndexV3) UnmarshalBinary(data []byte) error {
	i := &repo.IndexFile{}
	if err := yaml.Unmarshal(data, i); err != nil {
		return err
	}
	i.SortEntries()

	*idx = IndexV3{index: i}
	return nil
}

func (idx *IndexV3) Reader() (io.Reader, error) {
	b, err := idx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

func (idx *IndexV3) WriteFile(dest string, mode os.FileMode) error {
	return idx.index.WriteFile(dest, mode)
}

func newIndexV3() *IndexV3 {
	return &IndexV3{index: repo.NewIndexFile()}
}

func loadIndexV3(fpath string) (*IndexV3, error) {
	idx, err := repo.LoadIndexFile(fpath)
	if err != nil {
		return nil, err
	}
	return &IndexV3{index: idx}, nil
}
