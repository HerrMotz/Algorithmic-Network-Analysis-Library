package graphlib

import (
	"math/rand/v2"
	"reflect"
	"testing"
)

func TestLookupTable(t *testing.T) {
	names := []string{"berlin", "hamburg", "munich", "cologne"}
	lut, _ := InitializeLookupTable(len(names), names)
	if idx := lut["munich"]; idx != 2 {
		t.Errorf("expect munich → 2, got %d", idx)
	}
}

func TestAdjMatrix_HasEdge(t *testing.T) {
	m, err := InitializeAdjMat(4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = m.AddEdge(Edge{Vertex{0}, Vertex{1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = m.AddEdge(Edge{Vertex{1}, Vertex{2}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m.HasEdge(Edge{Vertex{0}, Vertex{1}}) {
		t.Errorf("expected edge (0,1) to exist")
	}
	if m.HasEdge(Edge{Vertex{2}, Vertex{3}}) {
		t.Errorf("edge (2,3) should not exist")
	}
}

// ----------------------------
// Correctness checks For Adjacency List
// ----------------------------

func TestAdjList_InternalConsistency(t *testing.T) {
	adj, _ := InitializeAdjList(4)
	if err := adj.AddEdge(Edge{Vertex{1}, Vertex{3}}); err != nil {
		return
	}
	if len(adj.List[1]) != 1 || len(adj.List[3]) != 1 {
		t.Fatalf("edge (1,3) not inserted symmetrically")
	}
}

func TestAdjList_HasArc(t *testing.T) {
	adj, _ := InitializeAdjList(4)
	if err := adj.AddArc(Arc{Vertex{1}, Vertex{3}}); err != nil {
		return
	}
	if !adj.HasArc(Arc{Vertex{1}, Vertex{3}}) {
		t.Errorf("expected arc (1,3) to exist")
	}
	if adj.HasArc(Arc{Vertex{2}, Vertex{3}}) {
		t.Errorf("arc (2,3) should not exist")
	}
}

func TestAdjList_HasEdge(t *testing.T) {
	adj, _ := InitializeAdjList(4)
	if err := adj.AddArc(Arc{Vertex{1}, Vertex{3}}); err != nil {
		return
	}
	if err := adj.AddEdge(Edge{Vertex{1}, Vertex{3}}); err != nil {
		return
	}
	if !adj.HasEdge(Edge{Vertex{1}, Vertex{3}}) {
		t.Errorf("expected edge (1,3) to exist")
	}
}

func TestAdjList_Neighbours(t *testing.T) {
	adj, _ := InitializeAdjList(4)
	// add some edges and test
	edges := []Edge{
		{Vertex{1}, Vertex{3}},
		{Vertex{1}, Vertex{2}},
		{Vertex{2}, Vertex{3}},
	}
	for _, e := range edges {
		mustAddEdge(t, adj, e)
	}

	// test neighbours
	neighbours, err := adj.Neighbours(Vertex{1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(neighbours) != 2 {
		t.Fatalf("expected 2 neighbours, got %d", len(neighbours))
	}
	_, ok1 := neighbours[2]
	_, ok2 := neighbours[3]
	if !ok1 || !ok2 {
		t.Fatalf("expected neighbours to be [2,3], got %v", neighbours)
	}
}

func mustAddEdge(t *testing.T, adj *AdjList, e Edge) {
	t.Helper()
	if err := adj.AddEdge(e); err != nil {
		t.Fatalf("AddEdge(%v) failed: %v", e, err)
	}
}

func TestShortestPaths_DAG(t *testing.T) {
	adj, _ := InitializeAdjList(4)

	zero := Vertex{0}
	one := Vertex{1}
	two := Vertex{2}
	three := Vertex{3}

	edges := []Edge{
		{zero, one},
		{zero, two},
		{one, three},
		{two, three},
	}

	for _, e := range edges {
		mustAddEdge(t, adj, e)
	}

	parents, distances := adj.DirectedAcyclicGraph(&zero)
	if len(parents) != 4 || len(distances) != 4 {
		t.Fatalf("expected 4 parents and distances, got %d and %d", len(parents), len(distances))
	}

	if parents[3] != two && parents[3] != one {
		t.Fatalf("expected parent of 3 to be 2 or 1, got %v", parents[3])
	}

	if distances[3] != 2 {
		t.Fatalf("expected distance to 3 to be 2, got %d", distances[3])
	}

	if parents[2] != zero {
		t.Fatalf("expected parent of 2 to be 0, got %v", parents[2])
	}

	if distances[2] != 1 {
		t.Fatalf("expected distance to 2 to be 1, got %d", distances[2])
	}
}

func TestSingleSourceShortestPaths(t *testing.T) {
	adj, _ := InitializeAdjList(4)

	zero := Vertex{0}
	one := Vertex{1}
	two := Vertex{2}
	three := Vertex{3}

	if err := adj.AddEdge(Edge{zero, one}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := adj.AddEdge(Edge{zero, two}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := adj.AddEdge(Edge{one, three}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := adj.AddEdge(Edge{two, three}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	distances, _ := adj.SingleSourceShortestPath(zero)
	if len(distances) != 4 {
		t.Fatalf("expected 4 distances, got %d", len(distances))
	}
	if distances[3] != 2 {
		t.Fatalf("expected distance to 3 to be 2, got %d", distances[3])
	}

	if distances[2] != 1 {
		t.Fatalf("expected distance to 2 to be 1, got %d", distances[2])
	}
}

func TestQueue(t *testing.T) {
	const capacity = 10

	queue := NewQueue[int](capacity)
	if queue.empty() == false {
		t.Errorf("expected queue to be empty")
	}

	const c1 = 12392
	const c2 = 9080890
	const c3 = 3748998
	const c4 = 546785

	queue.enqueue(c1)
	queue.enqueue(c2)

	if p1 := queue.pop(); p1 != c1 {
		t.Errorf("expected popped value to be c1, got %v", p1)
	}

	if p2 := queue.pop(); p2 != c2 {
		t.Errorf("expected popped value to be c2, got %v", p2)
	}

	if queue.empty() == false {
		t.Errorf("expected queue to be empty after popping to pushed values, got size %v", queue.size)
	}

	queue.enqueue(c3)

	if queue.empty() == true {
		t.Errorf("expected queue to not be after pushing one value")
	}

	if queue.size != 1 {
		t.Errorf("expected queue length to be 1, got %v", queue.size)
	}

	if queue.peek() != c3 {
		t.Errorf("expected peek to return current element, got %v", queue.size)
	}

	queue.enqueue(c4)
	if queue.size != 2 {
		t.Errorf("expected queue length to be 2, got %v", queue.size)
	}

	queue.grow()
	if queue.size != 2 {
		t.Errorf("expected queue length to be the same (2) after growing, got %v", queue.size)
	}

	queue.pop()
	queue.pop()

	const forceGrowCapacity = 3 * capacity

	list := make([]int, forceGrowCapacity)
	for i := 0; i < forceGrowCapacity; i++ {
		list[i] = rand.IntN(100 * capacity)
		queue.enqueue(list[i])
	}

	if queue.size != forceGrowCapacity {
		t.Errorf("expected queue length to be 2 + a constant \"force grow capacity\" %v, got %v", forceGrowCapacity, queue.size)
	}

	for i := 0; i < forceGrowCapacity; i++ {
		if _e := queue.pop(); _e != list[i] {
			t.Errorf("expected queue to reconstruct the input list order. at pos %v expected %v but got %v", i, list[i], _e)
		}
	}
}

func TestConnectedComponentsEdgeCases(t *testing.T) {
	adjOneNode, _ := InitializeAdjList(1)
	const expect = 0
	v := adjOneNode.NaiveDiameter()

	if expect != v {
		t.Errorf("Expected %v, but got %v", expect, v)
	}

	adjTwoNode, _ := InitializeAdjList(2)
	const expect2 = INF
	v = adjTwoNode.NaiveDiameter()

	if expect2 != v {
		t.Errorf("Expected %v, but got %v", expect2, v)
	}
}

func TestConnectedComponents(t *testing.T) {
	adj, _ := InitializeAdjList(8)

	vertices := [5]Vertex{
		{0},
		{1},
		{2},
		{3},
		{4},
	}

	edges := [4]Edge{
		{vertices[0], vertices[1]},
		{vertices[1], vertices[2]},
		{vertices[0], vertices[3]},
		{vertices[2], vertices[4]},
	}

	for _, e := range edges {
		mustAddEdge(t, adj, e)
	}

	component := adj.naiveConnectedComponents()
	expect := [8]int{0, 0, 0, 0, 0, 1, 2, 3}

	for i, v := range component {
		if expect[i] != v {
			t.Errorf("Expected %v, but got %v", expect[i], v)
		}
	}
}

func TestDiameter(t *testing.T) {
	adj, _ := InitializeAdjList(8)

	vertices := [8]Vertex{
		{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7},
	}

	edges := [4]Edge{
		{vertices[0], vertices[1]},
		{vertices[1], vertices[2]},
		{vertices[2], vertices[3]},
		{vertices[3], vertices[4]},
	}

	for _, edge := range edges {
		err := adj.AddEdge(edge)
		if err != nil {
			return
		}
	}

	diameter := adj.NaiveDiameter()
	diameter2 := adj.ImprovedDiameter()

	const expect = INF

	if diameter != expect {
		t.Errorf("Expected %v, but got %v", expect, diameter)
	}

	if diameter2 != expect {
		t.Errorf("Expected %v, but got %v", expect, diameter2)
	}

	moreEdges := [3]Edge{
		{vertices[4], vertices[5]},
		{vertices[5], vertices[6]},
		{vertices[6], vertices[7]},
	}

	for _, edge := range moreEdges {
		err := adj.AddEdge(edge)
		if err != nil {
			return
		}
	}

	const expect2 = 7

	diameter = adj.NaiveDiameter()
	diameter2 = adj.ImprovedDiameter()

	if diameter != expect2 {
		t.Errorf("Expected %v, got %v", expect2, diameter)
	}

	if diameter2 != expect2 {
		t.Errorf("Expected %v, got %v", expect2, diameter2)
	}
}

var exAdj0 = []map[int]Naught{
	{1: Naught{}},
	{0: Naught{}},
}
var wantCut0 = map[int]Naught{}
var wantNbh0 = map[int]Naught{
	0: {}, // this is the correct output only because the for loop starts at idx 0 in NeighbourhoodSubsets
}
var wantDist0 = [][]int{
	{0, 1},
	{1, 0},
}

var exAdj1 = []map[int]Naught{
	{1: Naught{}, 2: Naught{}},
	{0: Naught{}},
	{0: Naught{}},
	{},
}

var wantCut1 = map[int]Naught{0: {}}
var wantNbh1 = map[int]Naught{
	0: {},
	1: {}, // this is the correct output only because the for loop starts at idx 0 in NeighbourhoodSubsets
}
var wantDist1 = [][]int{
	{0, 1, 1, INF},
	{1, 0, 2, INF},
	{1, 2, 0, INF},
	{INF, INF, INF, 0},
}

var exAdj2 = []map[int]Naught{
	{3: Naught{}, 5: Naught{}},
	{2: Naught{}, 4: Naught{}},
	{1: Naught{}, 4: Naught{}},
	{0: Naught{}, 4: Naught{}},
	{1: Naught{}, 2: Naught{}, 3: Naught{}},
	{0: Naught{}},
}

var wantCut2 = map[int]Naught{0: {}, 3: {}, 4: {}}
var wantNbh2 = map[int]Naught{
	0: {},
	1: {},
	// 2 not, because 1 and 2 have the same neighbourhood
	3: {},
	4: {},
}
var wantDist2 = [][]int{
	{0, 3, 3, 1, 2, 1},
	{3, 0, 1, 2, 1, 4},
	{3, 1, 0, 2, 1, 4},
	{1, 2, 2, 0, 1, 2},
	{2, 1, 1, 1, 0, 3},
	{1, 4, 4, 2, 3, 0},
}

func TestCutVertices(t *testing.T) {
	tests := []struct {
		name string
		adj  AdjList
		size int
		want map[int]Naught
	}{
		{"tiny cycle", AdjList{List: exAdj0, Size: len(exAdj0)}, 2, wantCut0},
		{"star plus isolated", AdjList{List: exAdj1, Size: len(exAdj1)}, 4, wantCut1},
		{"more complex", AdjList{List: exAdj2, Size: len(exAdj2)}, 6, wantCut2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := tc.adj.CutVertices(tc.size)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("CutVertices(adj, %d) = %v; want %v", tc.size, got, tc.want)
			}
		})
	}
}

func TestNeighbourhoodSubsets(t *testing.T) {
	tests := []struct {
		name string
		adj  AdjList
		size int
		want map[int]Naught
	}{
		{"tiny cycle", AdjList{List: exAdj0, Size: len(exAdj0)}, 2, wantNbh0},
		{"star plus isolated", AdjList{List: exAdj1, Size: len(exAdj1)}, 4, wantNbh1},
		{"more complex", AdjList{List: exAdj2, Size: len(exAdj2)}, 6, wantNbh2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := tc.adj.NeighbourhoodSubsets()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NeighbourhoodSubsets(adj, %d) = %v; want %v", tc.size, got, tc.want)
			}
		})
	}
}

func TestImprovedAllPairsShortestPaths(t *testing.T) {
	tests := []struct {
		name string
		adj  AdjList
		size int
		want [][]int
	}{
		{"tiny cycle", AdjList{List: exAdj0, Size: len(exAdj0)}, 2, wantDist0},
		{"star plus isolated", AdjList{List: exAdj1, Size: len(exAdj1)}, 4, wantDist1},
		{"more complex", AdjList{List: exAdj2, Size: len(exAdj2)}, 6, wantDist2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := tc.adj.ImprovedAllPairsShortestPaths()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NeighbourhoodSubsets(adj, %d) = %v; want %v", tc.size, got, tc.want)
			}
		})
	}
}
