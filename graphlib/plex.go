package graphlib

import (
	"container/heap"
	"fmt"
	"sync"
)

// Copy Helper: Deep copy AdjList
func (l *AdjList) Copy() *AdjList {
	newList := make([]VertexSet, len(l.List))
	for i, m := range l.List {
		newMap := make(VertexSet, len(m))
		for k, v := range m {
			newMap[k] = v
		}
		newList[i] = newMap
	}
	return &AdjList{
		Size: l.Size,
		List: newList,
	}
}

// RemoveVertexCopy Helper: Remove a vertex and return a new graph
func (l *AdjList) RemoveVertexCopy(v Vertex) *AdjList {
	newL := l.Copy()
	err := newL.RemoveVertex(v)
	if err != nil {
		return nil
	}
	newL.List[v.Id] = nil
	return newL
}

// FirstBestDegree returns the first node with a degree lower than d or returns 0 and true
// if none of the nodes has degree <= d, then the Graph is a 2-plex
func (iGraph *InducedGraph) FirstBestDegree(d int) (Vertex, bool) {
	for i := range iGraph.GoodVertices {
		cmp, err := iGraph.Degree(Vertex{i})

		if err != nil {
			continue
		}
		if cmp < d {
			return Vertex{i}, false
		}
	}
	return Vertex{0}, true // the graph is a 2-plex
}

type VertexDegree struct {
	Vertex Vertex // your Vertex type
	Degree int
}

// PriorityQueue Is not a real queue or priority queue even, you
// have to use the heap function builtin to Go.
type PriorityQueue []VertexDegree

func (pq *PriorityQueue) Len() int           { return len(*pq) }
func (pq *PriorityQueue) Less(i, j int) bool { return (*pq)[i].Degree < (*pq)[j].Degree }
func (pq *PriorityQueue) Swap(i, j int)      { (*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i] }
func (pq *PriorityQueue) Push(x interface{}) { *pq = append(*pq, x.(VertexDegree)) }
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// MinBestDegree returns the node with the minimal degree < d,
// or (0, true) if all nodes have degree >= d (i.e., the graph is a 2-plex)
func (iGraph *InducedGraph) MinBestDegree(d int) (Vertex, bool) {
	pq := make(PriorityQueue, 0, iGraph.GoodVertices.Size())

	// Build the priority queue
	for i := range iGraph.GoodVertices {
		deg, err := iGraph.Degree(Vertex{i})
		if err != nil {
			continue // skip nodes with error
		}
		pq = append(pq, VertexDegree{Vertex: Vertex{i}, Degree: deg})
	}
	if len(pq) == 0 {
		return Vertex{0}, true
	}

	heap.Init(&pq)

	// Extract the minimal-degree node < d
	for pq.Len() > 0 {
		vd := heap.Pop(&pq).(VertexDegree)
		if vd.Degree < d {
			return vd.Vertex, false
		}
	}
	// No node with degree < d found
	return noV, true
}

// NotAdjacent returns two vertices not adjacent to u and not u
func (iGraph *InducedGraph) NotAdjacent(u Vertex) (Vertex, Vertex, error) {
	neigh, err := iGraph.Neighbours(u)
	if err != nil {
		return Vertex{-1}, Vertex{-1}, fmt.Errorf("failed to get neighbours of vertex %v: %w", u, err)
	}
	found := make([]int, 0, 2)
	for v := range iGraph.GoodVertices {
		if v == u.Id {
			continue // skip the vertex itself
		}
		if _, ok := neigh[v]; !ok {
			found = append(found, v)
			if len(found) == 2 {
				return Vertex{found[0]}, Vertex{found[1]}, nil
			}
		}
	}
	return Vertex{-1}, Vertex{-1}, fmt.Errorf("not enough non-neighbors")
}

// NotAdjacentWithDistanceTwo returns two vertices not adjacent to u and not u
// and have a maximal distance of 2 to u
func (iGraph *InducedGraph) NotAdjacentWithDistanceTwo(u Vertex) (Vertex, Vertex, error) {
	neigh1, err := iGraph.Neighbours(u)
	if err != nil {
		return Vertex{-1}, Vertex{-1}, fmt.Errorf("failed to get neighbours of vertex %v: %w", u, err)
	}

	visited := make(map[int]bool, iGraph.Size())
	visited[u.Id] = true
	pq := make(PriorityQueue, 0, iGraph.Size()) // todo chance to add a better heuristic

	for n1 := range neigh1 {
		visited[n1] = true // distance 1 to u
	}

	for n1 := range visited { // you can iterate over the slice and change it, because the loop uses a copy
		neigh2, err := iGraph.Neighbours(Vertex{n1}) // distance 2 to u
		if err != nil {
			return noV, noV, fmt.Errorf("NotAdjWithDistanceTwo: failed to get neighbours of vertex %v: %w", Vertex{n1}, err)
		}
		for n2 := range neigh2 { // distance 2 to u
			if !visited[n2] { // they do not also have distance 1 or 0 to u
				deg, err := iGraph.Degree(Vertex{n2})
				if err != nil {
					return noV, noV, err
				}
				pq = append(pq, VertexDegree{Vertex: Vertex{n2}, Degree: deg}) // from these we want to select two nodes v and w which have minimal degree
			}
		}
	}

	heap.Init(&pq)

	found := make([]int, 0, 2) // return slice
	for i := 0; i < pq.Len(); i++ {
		vd := heap.Pop(&pq).(VertexDegree)
		found = append(found, vd.Vertex.Id)
		if len(found) == 2 {
			return Vertex{found[0]}, Vertex{found[1]}, nil
		}
	}
	if len(found) == 0 {
		return specialNoV, u, fmt.Errorf("not enough non-neighbours (isolated vertex, 2-plex of size 1)")
	} else {
		return u, Vertex{found[0]}, fmt.Errorf("not enough neighbors (2-plex of size 2)")
	}
}

// TwoPlexNaive implements the naive algorithm
func (iGraph *InducedGraph) TwoPlexNaive() (VertexSet, error) {
	n := iGraph.Size()
	u, isTwoPlex := iGraph.FirstBestDegree(n - 2)
	if isTwoPlex {
		return iGraph.GoodVertices, nil
	}

	v, w, err := iGraph.NotAdjacent(u)
	if err != nil {
		// this can usually not occur, else it would have been a 2-plex
		return nil, err
	}

	// we need three new copies of InducedGraph, but the
	// reference to AdjList must stay the same so it does not get copied

	H引くu, err1A := iGraph.CreateReduced(u) // add no permanent vertex to set
	R1, err1B := H引くu.TwoPlexNaive()
	if err1A != nil || err1B != nil {
		return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v), TwoPlexPermanent(%v)", err1A, err1B)
	}

	H引くv, err2A := iGraph.CreateReduced(v) // add u as permanent vertex
	R2, err2B := H引くv.TwoPlexNaive()
	if err2A != nil || err2B != nil {
		return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v), TwoPlexPermanent(%v)", err2A, err2B)
	}

	H引くw, err3A := iGraph.CreateReduced(w) // add u and v as permanent vertices
	R3, err3B := H引くw.TwoPlexNaive()
	if err3A != nil || err3B != nil {
		return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v), TwoPlexPermanent(%v)", err3A, err3B)
	}

	// Find the largest node count
	results := []VertexSet{R1, R2, R3}
	maxIdx := 0
	maxNodes := len(results[0])
	for i := 1; i < 3; i++ {
		nodes := results[i].Size()
		if nodes > maxNodes {
			maxIdx = i
			maxNodes = nodes
		}
	}
	return results[maxIdx], nil
}

// TwoPlexNaiveMinDegree implements the naive algorithm but chooses vertex u with minimal degree
func (iGraph *InducedGraph) TwoPlexNaiveMinDegree() (VertexSet, error) {
	n := iGraph.Size()
	u, isTwoPlex := iGraph.MinBestDegree(n - 2)
	if isTwoPlex {
		return iGraph.GoodVertices, nil
	}

	v, w, err := iGraph.NotAdjacent(u)
	if err != nil {
		// this can usually not occur, else it would have been a 2-plex
		return nil, err
	}

	// we need three new copies of InducedGraph, but the
	// reference to AdjList must stay the same so it does not get copied

	H引くu, err1A := iGraph.CreateReduced(u) // add no permanent vertex to set
	R1, err1B := H引くu.TwoPlexNaiveMinDegree()
	if err1A != nil || err1B != nil {
		return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v), TwoPlexPermanent(%v)", err1A, err1B)
	}

	H引くv, err2A := iGraph.CreateReduced(v) // add u as permanent vertex
	R2, err2B := H引くv.TwoPlexNaiveMinDegree()
	if err2A != nil || err2B != nil {
		return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v), TwoPlexPermanent(%v)", err2A, err2B)
	}

	H引くw, err3A := iGraph.CreateReduced(w) // add u and v as permanent vertices
	R3, err3B := H引くw.TwoPlexNaiveMinDegree()
	if err3A != nil || err3B != nil {
		return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v), TwoPlexPermanent(%v)", err3A, err3B)
	}

	// Find the largest node count
	results := []VertexSet{R1, R2, R3}
	maxIdx := 0
	maxNodes := len(results[0])
	for i := 1; i < 3; i++ {
		nodes := results[i].Size()
		if nodes > maxNodes {
			maxIdx = i
			maxNodes = nodes
		}
	}
	return results[maxIdx], nil
}

// TwoPlexNaiveParallel explores the three recursive branches concurrently.
// depth is the current recursion level; it lets us stop spawning new
// goroutines once we are “deep enough” to keep the goroutine count in check.
func (iGraph *InducedGraph) TwoPlexNaiveParallel() (VertexSet, error) {
	n := iGraph.Size()
	u, isTwoPlex := iGraph.FirstBestDegree(n - 2)
	if isTwoPlex {
		return iGraph.GoodVertices, nil
	}

	v, w, err := iGraph.NotAdjacent(u)
	if err != nil {
		return nil, err
	}

	// Results struct for each branch
	type result struct {
		set        VertexSet
		errA, errB error
	}

	var wg sync.WaitGroup
	results := make([]result, 3)

	// Launch parallel recursive calls
	wg.Add(3)
	go func() {
		defer wg.Done()
		H引くu, errA := iGraph.CreateReduced(u)
		set, errB := H引くu.TwoPlexNaiveParallel()
		results[0] = result{set, errA, errB}
	}()
	go func() {
		defer wg.Done()
		H引くv, errA := iGraph.CreateReduced(v)
		set, errB := H引くv.TwoPlexNaiveParallel()
		results[1] = result{set, errA, errB}
	}()
	go func() {
		defer wg.Done()
		H引くw, errA := iGraph.CreateReduced(w)
		set, errB := H引くw.TwoPlexNaiveParallel()
		results[2] = result{set, errA, errB}
	}()
	wg.Wait()

	// Error checking
	for i, r := range results {
		if r.errA != nil || r.errB != nil {
			return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v), TwoPlexPermanent(%v) at branch %d", r.errA, r.errB, i)
		}
	}

	// Find the largest node count
	maxIdx := 0
	maxNodes := results[0].set.Size()
	for i := 1; i < 3; i++ {
		nodes := results[i].set.Size()
		if nodes > maxNodes {
			maxIdx = i
			maxNodes = nodes
		}
	}
	return results[maxIdx].set, nil
}

func _runTwoPlexIfPermanentHelper(pGraph *PermanentInducedGraph, testVertex Vertex, p1 Vertex, p2 Vertex) (VertexSet, error) {
	if !pGraph.PermanentVertices.HasVertex(testVertex) {
		reduced, errA := pGraph.CreateReduced(testVertex, p1, p2)
		if errA != nil {
			return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v)", errA)
		}
		result, errB := reduced.TwoPlexPermanent()
		if errB != nil {
			return nil, fmt.Errorf("failed to make recursion step: TwoPlexPermanent(%v)", errB)
		}
		return result, nil
	}
	return nil, nil
}

var noV = Vertex{-1}        // use -1 to indicate no vertex
var specialNoV = Vertex{-2} // used to indicate isolated node in distance function

// TODO
//  - Add distance improvement (remove all vertices that have distance >= 3 to a permanent vertex)

// TwoPlexPermanent implements the algorithm with permanent set
func (pGraph *PermanentInducedGraph) TwoPlexPermanent() (VertexSet, error) {
	n := pGraph.IGraph.Size()
	u, isTwoPlex := pGraph.IGraph.MinBestDegree(n - 2)
	if isTwoPlex {
		return pGraph.IGraph.GoodVertices, nil
	}

	v, w, err := pGraph.IGraph.NotAdjacent(u)
	if err != nil {
		// this can usually not occur, else it would have been a 2-plex
		return nil, err
	}

	// special case: if u, v, w are all in the permanent vertex set,
	// then we can return a trivial 2-Plex. If there is a larger 2-Plex,
	// it will be found in the other recursive calls.
	if pGraph.PermanentVertices.HasVertex(u) &&
		pGraph.PermanentVertices.HasVertex(v) &&
		pGraph.PermanentVertices.HasVertex(w) {
		return VertexSet{0: Naught{}, 1: Naught{}}, nil
	}

	var results = make([]VertexSet, 0, 3)

	// only call to u if it is not in the permanent vertex set
	if r, err := _runTwoPlexIfPermanentHelper(pGraph, u, noV, noV); err != nil { // noV means "no vertex"
		return nil, fmt.Errorf("failed to make recursion step: _runTwoPlexIfPermanentHelper(%v, %v, %v): %w", u, noV, noV, err)
	} else if r != nil {
		results = append(results, r)
	}

	// only call to v if it is not in the permanent vertex set
	if r, err := _runTwoPlexIfPermanentHelper(pGraph, v, u, noV); err != nil {
		return nil, fmt.Errorf("failed to make recursion step: _runTwoPlexIfPermanentHelper(%v, %v, %v): %w", v, u, noV, err)
	} else if r != nil {
		results = append(results, r)
	}

	// same for w
	if r, err := _runTwoPlexIfPermanentHelper(pGraph, w, u, v); err != nil {
		return nil, fmt.Errorf("failed to make recursion step: _runTwoPlexIfPermanentHelper(%v, %v, %v): %w", w, u, v, err)
	} else if r != nil {
		results = append(results, r)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no 2-Plex found")
	}

	// Find the largest node count
	maxIdx := 0
	maxNodes := len(results[0])
	for i := 1; i < len(results); i++ {
		nodes := results[i].Size()
		if nodes > maxNodes {
			maxIdx = i
			maxNodes = nodes
		}
	}
	return results[maxIdx], nil
}

func (pGraph *PermanentInducedGraph) TwoPlexPermanentParallelKaputt() (VertexSet, error) {
	n := pGraph.IGraph.Size()
	u, isTwoPlex := pGraph.IGraph.MinBestDegree(n - 2)
	if isTwoPlex {
		return pGraph.IGraph.GoodVertices, nil
	}

	v, w, err := pGraph.IGraph.NotAdjacent(u)
	if err != nil {
		return nil, err
	}

	if pGraph.PermanentVertices.HasVertex(u) &&
		pGraph.PermanentVertices.HasVertex(v) &&
		pGraph.PermanentVertices.HasVertex(w) {
		return VertexSet{0: Naught{}, 1: Naught{}}, nil
	}

	type plexResult struct {
		result VertexSet
		err    error
	}

	resultCh := make(chan plexResult, 3)

	calls := []struct {
		vertex, block1, block2 Vertex
	}{
		{u, noV, noV},
		{v, u, noV},
		{w, u, v},
	}

	launchCount := 0
	for _, call := range calls {
		if !pGraph.PermanentVertices.HasVertex(call.vertex) {
			// this does the same as _runTwoPlexIfDistanceHelper,
			// but we want the information whether the branch is going to be launched,
			// and we can reduce one call by expliciting it here.
			launchCount++
			go func(vertex, block1, block2 Vertex) {
				reduced, errA := pGraph.CreateReduced(vertex, block1, block2)
				if errA != nil {
					resultCh <- plexResult{result: nil, err: errA}
				}
				r, errB := reduced.TwoPlexPermanent()
				if errB != nil {
					resultCh <- plexResult{result: nil, err: errA}
				}
				resultCh <- plexResult{result: r, err: err}
			}(call.vertex, call.block1, call.block2)
		}
	}

	var maxPlex VertexSet

	for i := 0; i < launchCount; i++ {
		res := <-resultCh
		// return the first error that occurs
		if res.err != nil {
			return nil, fmt.Errorf("failed to make recursion step: %w", res.err)
		}
		if res.result != nil {
			size := res.result.Size()
			if maxPlex == nil || size > maxPlex.Size() {
				maxPlex = res.result
			}
		}
	}
	if maxPlex != nil {
		return maxPlex, nil
	}
	close(resultCh) // Not strictly necessary here, but good practice
	return nil, fmt.Errorf("no 2-Plex found")
}

func _runTwoPlexIfDistanceHelper(pGraph *PermanentInducedGraph, vToRemove Vertex, p1 Vertex, p2 Vertex) (VertexSet, error) {
	if !pGraph.PermanentVertices.HasVertex(vToRemove) {
		// ↓ the following line is the only change, the rest is the same as in _runTwoPlexIfPermanentHelper
		reduced, errA := pGraph.CreateReducedWithDistance(vToRemove, p1, p2)
		// ↑ changed line!
		if errA != nil {
			return nil, fmt.Errorf("failed to make recursion step: CreateReducedWithDistance(%v, %v, %v): %w", vToRemove, p1, p2, errA)
		}
		result, errB := reduced.TwoPlexDistance() // oh I forgot - this one changed as well.
		if errB != nil {
			return nil, fmt.Errorf("failed to make recursion step: TwoPlexDistance(%v)", errB)
		}
		return result, nil
	}
	return nil, nil
}

// TwoPlexDistance implement the algorithm with permanent set and distance improvement
func (pGraph *PermanentInducedGraph) TwoPlexDistance() (VertexSet, error) {
	n := pGraph.IGraph.Size()
	u, isTwoPlex := pGraph.IGraph.MinBestDegree(n - 2)
	if isTwoPlex {
		return pGraph.IGraph.GoodVertices, nil
	}

	var results = make([]VertexSet, 0, 3)

	// branch on u even if there are not two vertices with distance two to it
	// only call to u if it is not in the permanent vertex set
	if r, err := _runTwoPlexIfDistanceHelper(pGraph, u, noV, noV); err != nil { // noV means "no vertex"
		return nil, fmt.Errorf("failed to make recursion step: _runTwoPlexIfDistanceHelper(%v, %v, %v): %w", u, noV, noV, err)
	} else if r != nil {
		results = append(results, r)
	}

	v, w, err := pGraph.IGraph.NotAdjacentWithDistanceTwo(u)
	if err != nil {
		// if NotAdjacentWithDistanceTwo(u) encounters an error (it couldn't find
		// two vertices v and w), return the 2-plex from the H-u branch.
		if len(results) > 0 {
			return results[0], nil
		}
		// otherwise, propagate the error as before.
		return nil, err
	}

	//old code that didn't return the result from the u-branch
	/*v, w, err := pGraph.IGraph.NotAdjacentWithDistanceTwo(u)
	if err != nil {
		if v == specialNoV { // isolated vertex, 2 plex of size 1
			return VertexSet{0: Naught{}}, nil
		} else if v.Id >= 0 && w.Id >= 0 { // 2 plex of size 2
			return VertexSet{v.Id: Naught{}, w.Id: Naught{}}, nil
		}
		return nil, err
	}*/

	// special case: if u, v, w are all in the permanent vertex set,
	// then we can return a trivial 2-Plex. If there is a larger 2-Plex,
	// it will be found in the other recursive calls.
	if pGraph.PermanentVertices.HasVertex(u) &&
		pGraph.PermanentVertices.HasVertex(v) &&
		pGraph.PermanentVertices.HasVertex(w) {
		return VertexSet{0: Naught{}, 1: Naught{}}, nil
	}

	// only call to v if it is not in the permanent vertex set
	if r, err := _runTwoPlexIfDistanceHelper(pGraph, v, u, noV); err != nil {
		return nil, fmt.Errorf("failed to make recursion step: _runTwoPlexIfDistanceHelper(%v, %v, %v): %w", v, u, noV, err)
	} else if r != nil {
		results = append(results, r)
	}

	// same for w
	if r, err := _runTwoPlexIfDistanceHelper(pGraph, w, u, v); err != nil {
		return nil, fmt.Errorf("failed to make recursion step: _runTwoPlexIfDistanceHelper(%v, %v, %v): %w", w, u, v, err)
	} else if r != nil {
		results = append(results, r)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no 2-Plex found")
	}

	// Find the largest node count
	maxIdx := 0
	maxNodes := len(results[0])
	for i := 1; i < len(results); i++ {
		nodes := results[i].Size()
		if nodes > maxNodes {
			maxIdx = i
			maxNodes = nodes
		}
	}
	return results[maxIdx], nil
}

// TwoPlexNStar implements the algorithm with storing the size (nStar) of the current biggest 2-plex found
// initializes nStar := 0 and calls the recursive helper.
func (iGraph *InducedGraph) TwoPlexNStar() (VertexSet, error) {
	// declare a local variable for nStar and initialize it to 0.
	initialNStar := 0
	// call the recursive helper function, passing the address of initialNStar.
	// The helper function will modify initialNStar directly.
	return iGraph.TwoPlexNStarRecursive(&initialNStar)
}

// TwoPlexNStarRecursive recursive function that performs the 2-plex finding
// takes a pointer to nStar, allowing it to be updated globally across the recursion tree.
// vertices with degree at most nStar - 2 are removed from the graph
// cancel branches with at most nStar vertices
func (iGraph *InducedGraph) TwoPlexNStarRecursive(nStarPtr *int) (VertexSet, error) { // nStarPtr is a pointer
	n := iGraph.Size() // initial size

	// initial size check:
	if n <= *nStarPtr {
		return VertexSet{}, nil // return an empty set, indicating no larger 2-plex found in this branch
	}

	// create a new VertexSet to hold the vertices that will remain in the graph after deleting these with a
	// degree <= nStar - 2 (building a new GoodVertices set)
	currentGoodVertices := iGraph.GoodVertices.Copy() // copy of the current good vertices

	// iterate through all vertices in the current GoodVertices
	for vertexID := range currentGoodVertices {
		// calculate the degree of the vertex in the current induced graph
		degree, err := iGraph.Degree(Vertex{vertexID})
		if err != nil {
			return nil, fmt.Errorf("error getting degree for vertex %d during simple pruning: %w", vertexID, err)
		}
		// if degree is too small, remove it.
		if degree < *nStarPtr-1 {
			// remove the vertex directly from the currentGoodVertices set
			delete(currentGoodVertices, vertexID)
		}
	}

	// update iGraph.GoodVertices to the new vertex set.
	iGraph.GoodVertices = currentGoodVertices

	n = iGraph.Size() // update graph size

	// re-check the graph size after deleting vertices
	if n <= *nStarPtr {
		return VertexSet{}, nil
	}

	u, isTwoPlex := iGraph.MinBestDegree(n - 2)
	if isTwoPlex {
		// if it's a 2-plex, check if it's larger than nStarPtr's value (it always should be...)
		if iGraph.GoodVertices.Size() > *nStarPtr {
			*nStarPtr = iGraph.GoodVertices.Size() // update global nStarPtr
		}
		return iGraph.GoodVertices, nil //return found 2-plex
	}

	v, w, err := iGraph.NotAdjacent(u)
	if err != nil {
		// this can usually not occur, else it would have been a 2-plex.
		return nil, err
	}

	var results []VertexSet

	// branch 1: exclude u
	H引くu, err1A := iGraph.CreateReduced(u) // add no permanent vertex to set
	if err1A != nil {
		return nil, fmt.Errorf("failed to make recursion step: CreateReduced(exclude %v): %w", u, err1A)
	}
	// only recurse if the reduced graph might contain a larger 2-plex than the current global nStar
	if H引くu.Size() > *nStarPtr { // check against global nStar
		R1, err1B := H引くu.TwoPlexNStarRecursive(nStarPtr) // pass the pointer down
		if err1B != nil {
			return nil, fmt.Errorf("failed to make recursion step: twoPlexNStarRecursive(exclude %v): %w", u, err1B)
		}
		if R1.Size() > 0 { // Only add if a valid 2-plex was found in this branch
			results = append(results, R1)
			// the global nStarPtr is updated inside the recursive call, but we re-check here for consistency if R1 itself is the new max.
			if R1.Size() > *nStarPtr { // This check is mostly for clarity; the child call should have updated it.
				*nStarPtr = R1.Size()
			}
		}
	}

	// branch 2: exclude v
	H引くv, err2A := iGraph.CreateReduced(v) // add u as permanent vertex
	if err2A != nil {
		return nil, fmt.Errorf("failed to make recursion step: CreateReduced(exclude %v, permanent %v): %w", v, u, err2A)
	}
	// only recurse if the reduced graph might contain a larger 2-plex than the current global nStar
	if H引くv.Size() > *nStarPtr { // check against global nStar
		R2, err2B := H引くv.TwoPlexNStarRecursive(nStarPtr)
		if err2B != nil {
			return nil, fmt.Errorf("failed to make recursion step: twoPlexNStarRecursive(exclude %v, permanent %v): %w", v, u, err2B)
		}
		if R2.Size() > 0 {
			results = append(results, R2)
			if R2.Size() > *nStarPtr {
				*nStarPtr = R2.Size()
			}
		}
	}

	// branch 3: exclude w
	H引くw, err3A := iGraph.CreateReduced(w) // Add u and v as permanent vertices (implies u and v are kept, w is excluded)
	if err3A != nil {
		return nil, fmt.Errorf("failed to make recursion step: CreateReduced(exclude %v, permanent %v, %v): %w", w, u, v, err3A)
	}
	// only recurse if the reduced graph might contain a larger 2-plex than the current global nStar
	if H引くw.Size() > *nStarPtr { // check against global nStar
		R3, err3B := H引くw.TwoPlexNStarRecursive(nStarPtr)
		if err3B != nil {
			return nil, fmt.Errorf("failed to make recursion step: twoPlexNStarRecursive(exclude %v, permanent %v, %v): %w", w, u, v, err3B)
		}
		if R3.Size() > 0 {
			results = append(results, R3)
			if R3.Size() > *nStarPtr {
				*nStarPtr = R3.Size()
			}
		}
	}

	if len(results) == 0 {
		return VertexSet{}, nil
	}

	// find the largest node count among the current branch's results
	maxIdx := 0
	maxNodes := results[0].Size()
	for i := 1; i < len(results); i++ {
		nodes := results[i].Size()
		if nodes > maxNodes {
			maxIdx = i
			maxNodes = nodes
		}
	}

	// after finding the local maximum, update the global nStar if this local max is larger
	if maxNodes > *nStarPtr {
		*nStarPtr = maxNodes
	}

	return results[maxIdx], nil
}
