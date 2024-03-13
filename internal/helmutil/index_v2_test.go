package helmutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
)

func TestIndexV2_MarshalBinary(t *testing.T) {
	idx := IndexV2{
		index: &repo.IndexFile{
			APIVersion: "foo",
			Generated:  time.Date(2018, 01, 01, 0, 0, 0, 0, time.UTC),
		},
	}

	b, err := idx.MarshalBinary()
	require.NoError(t, err)

	expected := `apiVersion: foo
entries: null
generated: "2018-01-01T00:00:00Z"
`
	assert.Equal(t, expected, string(b))
}

func TestIndexV2_UnmarshalBinary(t *testing.T) {
	input := []byte(`apiVersion: foo
entries: null
generated: 2018-01-01T00:00:00Z
`)

	idx := &IndexV2{}
	err := idx.UnmarshalBinary(input)
	require.NoError(t, err)

	assert.Equal(t, "foo", idx.index.APIVersion)
	assert.Equal(t, time.Date(2018, 01, 01, 0, 0, 0, 0, time.UTC), idx.index.Generated)
}

func TestIndexV2_AddOrReplace(t *testing.T) {
	t.Run("should add a new chart", func(t *testing.T) {
		i := newIndexV2()

		err := i.AddOrReplace(
			&chart.Metadata{
				Name:    "foo",
				Version: "0.1.0",
			},
			"foo-0.1.0.tgz",
			"http://example.com/charts",
			"sha256:1234567890",
		)
		require.NoError(t, err)

		assert.Equal(t, "http://example.com/charts/foo-0.1.0.tgz", i.index.Entries["foo"][0].URLs[0])
	})

	t.Run("should add a new version of a chart", func(t *testing.T) {
		i := newIndexV2()

		err := i.AddOrReplace(
			&chart.Metadata{
				Name:    "foo",
				Version: "0.1.0",
			},
			"foo-0.1.0.tgz",
			"http://example.com/charts",
			"sha256:111",
		)
		require.NoError(t, err)

		err = i.AddOrReplace(
			&chart.Metadata{
				Name:    "foo",
				Version: "0.1.1",
			},
			"foo-0.1.1.tgz",
			"http://example.com/charts",
			"sha256:222",
		)
		require.NoError(t, err)

		i.SortEntries()

		assert.Equal(t, "http://example.com/charts/foo-0.1.1.tgz", i.index.Entries["foo"][0].URLs[0])
		assert.Equal(t, "sha256:222", i.index.Entries["foo"][0].Digest)
	})

	t.Run("should replace existing chart version", func(t *testing.T) {
		i := newIndexV2()

		err := i.AddOrReplace(
			&chart.Metadata{
				Name:    "foo",
				Version: "0.1.0",
			},
			"foo-0.1.0.tgz",
			"http://example.com/charts",
			"sha256:111",
		)
		require.NoError(t, err)

		err = i.AddOrReplace(
			&chart.Metadata{
				Name:    "foo",
				Version: "0.1.0",
			},
			"foo-0.1.0.tgz",
			"http://example.com/charts",
			"sha256:222",
		)
		require.NoError(t, err)

		require.Len(t, i.index.Entries, 1)

		assert.Equal(t, "http://example.com/charts/foo-0.1.0.tgz", i.index.Entries["foo"][0].URLs[0])
		assert.Equal(t, "sha256:222", i.index.Entries["foo"][0].Digest)
	})
}

func TestIndexV2_UpdateGeneratedTime(t *testing.T) {
	idx := IndexV2{
		index: &repo.IndexFile{
			APIVersion: "foo",
			Generated:  time.Date(2018, 01, 01, 0, 0, 0, 0, time.UTC),
		},
	}
	generatedOld := idx.index.Generated
	idx.UpdateGeneratedTime()
	generatedNew := idx.index.Generated
	assert.True(t, generatedNew.After(generatedOld), "Expected %s greater than %s", generatedNew.String(), generatedOld.String())
}
