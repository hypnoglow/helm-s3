package helmutil

import (
	"encoding/json"
	"fmt"
	"io"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// ChartV3 implements Chart in Helm v3.
type ChartV3 struct {
	chart *chart.Chart
}

func (c ChartV3) Name() string {
	return c.chart.Name()
}

func (c ChartV3) Version() string {
	if c.chart.Metadata == nil {
		return ""
	}
	return c.chart.Metadata.Version
}

func (c ChartV3) Metadata() ChartMetadata {
	return &chartMetadataV3{meta: c.chart.Metadata}
}

func loadChartV3(fpath string) (ChartV3, error) {
	ch, err := loader.LoadFile(fpath)
	if err != nil {
		return ChartV3{}, fmt.Errorf("failed to load chart file: %s", err.Error())
	}
	return ChartV3{chart: ch}, nil
}

func loadArchiveV3(r io.Reader) (ChartV3, error) {
	ch, err := loader.LoadArchive(r)
	if err != nil {
		return ChartV3{}, fmt.Errorf("failed to load chart archive: %s", err.Error())
	}
	return ChartV3{chart: ch}, nil
}

type chartMetadataV3 struct {
	meta *chart.Metadata
}

func (c *chartMetadataV3) MarshalJSON() ([]byte, error) {
	if c.meta == nil {
		return nil, nil
	}
	return json.Marshal(c.meta)
}

func (c *chartMetadataV3) UnmarshalJSON(b []byte) error {
	if c.meta == nil {
		c.meta = &chart.Metadata{}
	}
	return json.Unmarshal(b, c.meta)
}

func (c *chartMetadataV3) Value() interface{} {
	return c.meta
}

func newChartMetadataV3() *chartMetadataV3 {
	return &chartMetadataV3{meta: &chart.Metadata{}}
}
