package graphlib

import (
	"reflect"
	"testing"
)

func TestGenerateRandomGraphIsReproducible(t *testing.T) {
	_, edges1, err := GenerateRandomGraph(60, 0.2, false, false, 42)
	if err != nil {
		t.Fatal(err)
	}
	_, edges2, err := GenerateRandomGraph(60, 0.2, false, false, 42)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(edges1, edges2) {
		t.Fatal("the same seed produced different graphs")
	}

	_, edges3, err := GenerateRandomGraph(60, 0.2, false, false, 43)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(edges1, edges3) {
		t.Fatal("different seeds produced the same graph")
	}
}
