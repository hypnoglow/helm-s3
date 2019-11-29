package helmutil

import (
	"encoding/json"
	"fmt"
	"io"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// ChartV2 implements Chart in Helm v2.
type ChartV2 struct {
	chart *chart.Chart
}

func (c ChartV2) Name() string {
	return c.chart.GetMetadata().GetName()
}

func (c ChartV2) Version() string {
	return c.chart.GetMetadata().GetVersion()
}

func (c ChartV2) Metadata() ChartMetadata {
	return &chartMetadataV2{meta: c.chart.GetMetadata()}
}

func loadChartV2(fpath string) (ChartV2, error) {
	ch, err := chartutil.LoadFile(fpath)
	if err != nil {
		return ChartV2{}, fmt.Errorf("failed to load chart file: %s", err.Error())
	}
	return ChartV2{chart: ch}, nil
}

func loadArchiveV2(r io.Reader) (ChartV2, error) {
	ch, err := chartutil.LoadArchive(r)
	if err != nil {
		return ChartV2{}, fmt.Errorf("failed to load chart archive: %s", err.Error())
	}
	return ChartV2{chart: ch}, nil
}

type chartMetadataV2 struct {
	meta *chart.Metadata
}

func (c *chartMetadataV2) MarshalJSON() ([]byte, error) {
	if c.meta == nil {
		return nil, nil
	}
	return json.Marshal(c.meta)
}

func (c *chartMetadataV2) UnmarshalJSON(b []byte) error {
	if c.meta == nil {
		c.meta = &chart.Metadata{}
	}
	return json.Unmarshal(b, c.meta)
}

func (c *chartMetadataV2) Value() interface{} {
	return c.meta
}

func newChartMetadataV2() *chartMetadataV2 {
	return &chartMetadataV2{meta: &chart.Metadata{}}
}
