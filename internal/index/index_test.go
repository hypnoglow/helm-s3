package index

import (
	"testing"
	"time"

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
generated: 2018-01-01T00:00:00Z
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
