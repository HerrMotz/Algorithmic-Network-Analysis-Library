package graphlib

import (
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
// TODO add a function which select minimal degree using a priority queue
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
		result, errB := reduced.TwoPlexPermanent()
		if errA != nil || errB != nil {
			return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v), TwoPlexPermanent(%v)", errA, errB)
		}
		return result, nil
	}
	return nil, nil
}

var noV = Vertex{-1} // use -1 to indicate no vertex

// TODO
//  - Add more tests for the TwoPlexPermanent algorithm (bigger graphs for performance testing)
//  - Add the minimal degree optimisation
//  - Add distance improvement (remove all vertices that have distance >= 3 to a permanent vertex)
//  - Store the size of the currently found 2-Plex (stop recursion if it is not larger than the current best)
//  - Remove nodes with at most degree of n*-2

// TwoPlexPermanent implements the algorithm with permanent set
func (pGraph *PermanentInducedGraph) TwoPlexPermanent() (VertexSet, error) {
	n := pGraph.IGraph.Size()
	u, isTwoPlex := pGraph.IGraph.FirstBestDegree(n - 2)
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

func (pGraph *PermanentInducedGraph) TwoPlexPermanentParallel() (VertexSet, error) {
	n := pGraph.IGraph.Size()
	u, isTwoPlex := pGraph.IGraph.FirstBestDegree(n - 2)
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

func _runTwoPlexIfDistanceHelper(pGraph *PermanentInducedGraph, testVertex Vertex, p1 Vertex, p2 Vertex) (VertexSet, error) {
	if !pGraph.PermanentVertices.HasVertex(testVertex) {
		// ↓ the following line is the only change, the rest is the same as in _runTwoPlexIfPermanentHelper
		reduced, errA := pGraph.CreateReducedWithDistance(testVertex, p1, p2)
		// ↑ changed line!
		result, errB := reduced.TwoPlexDistance() // oh I forgot - this one changed as well.
		if errA != nil || errB != nil {
			return nil, fmt.Errorf("failed to make recursion step: CreateReducedWithDistance(%v), TwoPlexDistance(%v)", errA, errB)
		}
		return result, nil
	}
	return nil, nil
}

// TwoPlexDistance implement the algorithm with permanent set and distance improvement
func (pGraph *PermanentInducedGraph) TwoPlexDistance() (VertexSet, error) {
	n := pGraph.IGraph.Size()
	u, isTwoPlex := pGraph.IGraph.FirstBestDegree(n - 2)
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
	if r, err := _runTwoPlexIfDistanceHelper(pGraph, u, noV, noV); err != nil { // noV means "no vertex"
		return nil, fmt.Errorf("failed to make recursion step: _runTwoPlexIfDistanceHelper(%v, %v, %v): %w", u, noV, noV, err)
	} else if r != nil {
		results = append(results, r)
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
