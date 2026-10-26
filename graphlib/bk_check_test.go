// bk_check_test.go -----------------------------------------
package graphlib

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"
)

func keyOf(C []int) string {
	tmp := append([]int(nil), C...) // Kopie, um Caller nicht zu verändern
	sort.Ints(tmp)
	return fmt.Sprint(tmp)
}

func loadTruth(path string) ([][]int, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var truth [][]int
	return truth, json.Unmarshal(f, &truth)
}

func TestBronKerboschVariants(t *testing.T) {
	g, err := ReadGraphAdjList("../exampleGraphs/soc-karate.mtx", true)
	if err != nil {
		t.Fatalf("graph: %v", err)
	}
	g.MakeUndirected()

	truth, err := loadTruth("../truths/truth.json")
	if err != nil {
		t.Fatalf("truth: %v", err)
	}

	want := make(map[string]struct{}, len(truth))
	for _, c := range truth {
		want[keyOf(c)] = struct{}{}
	}

	variants := []struct {
		name string
		run  func() [][]int
	}{
		{"SimpleIter", g.BronKerboschSimpleIter},
		{"Pivot", g.BronKerboschPivot},
		{"Degeneracy", g.BronKerboschDegeneracy},
	}

	for _, v := range variants {
		got := make(map[string]struct{})
		for _, c := range v.run() {
			got[keyOf(c)] = struct{}{}
		}

		// 1) Zuerst reine Anzahl vergleichen
		if len(got) != len(want) {
			t.Errorf("%s – Anzahl: got %d, want %d", v.name, len(got), len(want))
		}

		// 2) Danach nur noch Differenz‐Listing, falls gewünscht
		for k := range want {
			if _, ok := got[k]; !ok {
				t.Errorf("%s – fehlende Clique %s", v.name, k)
			}
		}
		for k := range got {
			if _, ok := want[k]; !ok {
				t.Errorf("%s – überzählige Clique %s", v.name, k)
			}
		}
	}
}
