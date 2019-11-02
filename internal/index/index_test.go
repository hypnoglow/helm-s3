package index

import (
	"testing"
	"time"

	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
)

func TestIndex_MarshalBinary(t *testing.T) {
	idx := Index{
		IndexFile: &repo.IndexFile{
			APIVersion: "foo",
			Generated:  time.Date(2018, 01, 01, 0, 0, 0, 0, time.UTC),
		},
	}

	b, err := idx.MarshalBinary()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := `apiVersion: foo
entries: null
generated: "2018-01-01T00:00:00Z"
`
	if string(b) != expected {
		t.Errorf("Expected %q but got %q", expected, string(b))
	}
}

func TestIndex_UnmarshalBinary(t *testing.T) {
	input := []byte(`apiVersion: foo
entries: null
generated: 2018-01-01T00:00:00Z
`)

	idx := &Index{}
	if err := idx.UnmarshalBinary(input); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedVersion := "foo"
	if idx.APIVersion != expectedVersion {
		t.Errorf("Expected %q but got %q", "foo", idx.APIVersion)
	}

	expectedDate := time.Date(2018, 01, 01, 0, 0, 0, 0, time.UTC)
	if idx.Generated != expectedDate {
		t.Errorf("Expected %q but got %q", expectedDate, idx.Generated)
	}
}

func TestIndex_AddOrReplace(t *testing.T) {
	t.Run("should add a new chart", func(t *testing.T) {
		i := New()
		i.AddOrReplace(
			&chart.Metadata{
				Name:    "foo",
				Version: "0.1.0",
			},
			"foo-0.1.0.tgz",
			"http://example.com/charts",
			"sha256:1234567890",
		)

		if i.Entries["foo"][0].URLs[0] != "http://example.com/charts/foo-0.1.0.tgz" {
			t.Errorf("Expected http://example.com/charts/foo-0.1.0.tgz, got %s", i.Entries["foo"][0].URLs[0])
		}
	})

	t.Run("should add a new version of a chart", func(t *testing.T) {
		i := New()
		i.AddOrReplace(
			&chart.Metadata{
				Name:    "foo",
				Version: "0.1.0",
			},
			"foo-0.1.0.tgz",
			"http://example.com/charts",
			"sha256:111",
		)
		i.AddOrReplace(
			&chart.Metadata{
				Name:    "foo",
				Version: "0.1.1",
			},
			"foo-0.1.1.tgz",
			"http://example.com/charts",
			"sha256:222",
		)
		i.SortEntries()

		if i.Entries["foo"][0].URLs[0] != "http://example.com/charts/foo-0.1.1.tgz" {
			t.Errorf("Expected http://example.com/charts/foo-0.1.1.tgz, got %s", i.Entries["foo"][0].URLs[0])
		}
		if i.Entries["foo"][0].Digest != "sha256:222" {
			t.Errorf("Expected sha256:222 but got %s", i.Entries["foo"][0].Digest)
		}
	})

	t.Run("should replace existing chart version", func(t *testing.T) {
		i := New()
		i.AddOrReplace(
			&chart.Metadata{
				Name:    "foo",
				Version: "0.1.0",
			},
			"foo-0.1.0.tgz",
			"http://example.com/charts",
			"sha256:111",
		)
		i.AddOrReplace(
			&chart.Metadata{
				Name:    "foo",
				Version: "0.1.0",
			},
			"foo-0.1.0.tgz",
			"http://example.com/charts",
			"sha256:222",
		)

		if len(i.Entries) != 1 {
			t.Fatalf("Expected 1 entry but got %d", len(i.Entries))
		}

		if i.Entries["foo"][0].URLs[0] != "http://example.com/charts/foo-0.1.0.tgz" {
			t.Errorf("Expected http://example.com/charts/foo-0.1.0.tgz, got %s", i.Entries["foo"][0].URLs[0])
		}
		if i.Entries["foo"][0].Digest != "sha256:222" {
			t.Errorf("Expected sha256:222 but got %s", i.Entries["foo"][0].Digest)
		}
	})
}
