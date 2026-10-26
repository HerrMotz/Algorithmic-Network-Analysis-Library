package graphlib

import (
	"bytes"
	"reflect"
	"testing"
)

func TestFprintClusters(t *testing.T) {
	clusters := []VertexSet{
		{4: Naught{}},
		{2: Naught{}, 0: Naught{}, 1: Naught{}},
		{3: Naught{}},
	}
	before := []VertexSet{clusters[0], clusters[1], clusters[2]}

	var buf bytes.Buffer
	FprintClusters(&buf, clusters)

	want := `CLUSTERING
========================================
C0 (|C|=3) : [0 1 2]
C1 (|C|=1) : [4]
C2 (|C|=1) : [3]
----------------------------------------
Total: 3 clusters
Size distribution:
   2× cluster of size 1
   1× cluster of size 3
========================================
`
	if got := buf.String(); got != want {
		t.Errorf("unexpected output:\n%s\nwant:\n%s", got, want)
	}
	if !reflect.DeepEqual(clusters, before) {
		t.Errorf("FprintClusters reordered its argument: %v", clusters)
	}

	buf.Reset()
	FprintClusters(&buf, nil)
	if got := buf.String(); got != "(no clusters)\n" {
		t.Errorf("empty clustering: got %q", got)
	}
}
