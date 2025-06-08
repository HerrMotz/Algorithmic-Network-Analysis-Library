package graphlib

import (
	"errors"
	"fmt"
)

type InducedGraph struct {
	AdjList      *AdjList
	GoodVertices VertexSet
}

type PermanentInducedGraph struct {
	IGraph            InducedGraph
	PermanentVertices VertexSet
}

func (iGraph *InducedGraph) Size() int {
	return iGraph.GoodVertices.Size()
}

func (iGraph *InducedGraph) RemoveVertex(v Vertex) error {
	return iGraph.MarkVertexAsBad(v)
}

func (iGraph *InducedGraph) MarkVertexAsBad(v Vertex) error {
	// marks vertex v as bad in the induced graph
	if iGraph.AdjList == nil || iGraph.GoodVertices == nil {
		return errors.New("induced graph is not initialized")
	}
	if !iGraph.GoodVertices.HasVertex(v) {
		return errors.New(fmt.Sprintf("Vertex %v does not exist", v))
	}

	// Remove vertex from GoodVertices
	iGraph.GoodVertices.RemoveVertex(v)

	// Remove vertex from AdjList
	iGraph.AdjList = iGraph.AdjList.RemoveVertexCopy(v)
	return nil
}

func (iGraph *InducedGraph) Neighbours(v Vertex) (VertexSet, error) {
	// returns the neighbours of vertex v in the induced graph
	if iGraph.AdjList == nil || iGraph.GoodVertices == nil {
		return nil, errors.New("induced graph is not initialized")
	}
	if !iGraph.GoodVertices.HasVertex(v) {
		return nil, errors.New(fmt.Sprintf("Vertex %v does not exist", v))
	}

	neighbours, err := iGraph.AdjList.Neighbours(v)
	if err != nil {
		return nil, err
	}
	// copy the neighbours to avoid modifying the original set
	neighbours = neighbours.Copy()

	// Filter neighbours to only include those that are in GoodVertices
	if iGraph.GoodVertices.Size() < iGraph.AdjList.Size { // only check if node set is smaller than the original graph
		for k := range neighbours {
			if !iGraph.GoodVertices.HasVertex(Vertex{k}) {
				neighbours.RemoveVertex(Vertex{k})
			}
		}
	}
	return neighbours, nil
}

func (iGraph *InducedGraph) HasVertex(v Vertex) bool {
	if iGraph.GoodVertices == nil {
		return false
	}
	return iGraph.GoodVertices.HasVertex(v)
}

func (iGraph *InducedGraph) Degree(v Vertex) (int, error) {
	// returns the degree of vertex v in the induced graph
	if iGraph.AdjList == nil || iGraph.GoodVertices == nil {
		return 0, errors.New("induced graph is not initialized")
	}
	if !iGraph.GoodVertices.HasVertex(v) {
		return 0, errors.New(fmt.Sprintf("Vertex %v does not exist", v))
	}

	neighbours, err := iGraph.Neighbours(v)
	if err != nil {
		return 0, err
	}
	return len(neighbours), nil
}

// CreateReduced returns a new InducedGraph that shares the same
// AdjList as igraph but whose GoodVertices set has been copied, filtered
// and with vToRemove removed. The original igraph is not modified.
func (iGraph *InducedGraph) CreateReduced(
	vToRemove Vertex,
) (*InducedGraph, error) {
	if iGraph.AdjList == nil || iGraph.GoodVertices == nil {
		return nil, errors.New("induced graph is not initialized")
	}
	if !iGraph.GoodVertices.HasVertex(vToRemove) {
		return nil, fmt.Errorf("vertex %v does not exist in good vertices", vToRemove)
	}

	// Copy the GoodVertices set and remove the specified vertex
	newGood := iGraph.GoodVertices.CopyAndRemoveVertex(vToRemove)

	// Return a brand-new InducedGraph sharing the same AdjList
	return &InducedGraph{
		AdjList:      iGraph.AdjList,
		GoodVertices: newGood,
	}, nil
}

// CreateReduced combines the necessary steps of PGraph and IGraph to create a new
// PermanentInducedGraph that shares the same AdjList as pgraph.IGraph but
// where vToRemove is removed from the GoodVertices set.
func (pGraph *PermanentInducedGraph) CreateReduced(vToRemove, v1, v2 Vertex) (*PermanentInducedGraph, error) {
	if pGraph.IGraph.AdjList == nil || pGraph.IGraph.GoodVertices == nil {
		return nil, errors.New("permanent induced graph is not initialized")
	}

	rGraph, err := pGraph.IGraph.CreateReduced(vToRemove)
	if err != nil {
		return nil, fmt.Errorf("failed to create reduced graph: %w", err)
	}

	// Create a new PermanentInducedGraph with the reduced InducedGraph
	newPGraph := PermanentInducedGraph{
		IGraph:            *rGraph,
		PermanentVertices: pGraph.PermanentVertices.Copy(),
	}

	if v1.Id >= 0 {
		// ignore the error if the vertex already exists
		_ = newPGraph.PermanentVertices.AddVertex(v1)
	}
	if v2.Id >= 0 {
		_ = newPGraph.PermanentVertices.AddVertex(v2)
	}

	return &newPGraph, nil
}

func (pGraph *PermanentInducedGraph) MarkBadWithBfs(source Vertex, markBadAfterDistance int) error {
	size := pGraph.IGraph.Size()
	dist := make([]int, size)
	for i := range dist {
		dist[i] = INF
	}
	dist[source.Id] = 0
	queue := NewQueue[int](10)
	visited := make([]bool, size)
	visited[source.Id] = true

	queue.enqueue(source.Id)

	for queue.size > 0 {
		v := queue.pop()
		// Iterate over all neighbours of v
		neighbours, err := pGraph.IGraph.Neighbours(Vertex{v})
		if err != nil {
			return err
		}

		for k := range neighbours {
			if !visited[k] {
				queue.enqueue(k)
				visited[k] = true
				dist[v] = dist[v] + 1
				// Mark the vertex as bad in the InducedGraph if the distance is larger than markBadAfterDistance
				if dist[v] > markBadAfterDistance {
					err := pGraph.IGraph.MarkVertexAsBad(Vertex{k})
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// CreateReducedWithDistance combined the distance improvement with the permanent set
func (pGraph *PermanentInducedGraph) CreateReducedWithDistance(vToRemove, v1, v2 Vertex) (*PermanentInducedGraph, error) {
	if pGraph.IGraph.AdjList == nil || pGraph.IGraph.GoodVertices == nil {
		return nil, errors.New("permanent induced graph is not initialized")
	}

	rGraph, err := pGraph.IGraph.CreateReduced(vToRemove)
	if err != nil {
		return nil, fmt.Errorf("failed to create reduced graph: %w", err)
	}

	// Create a new PermanentInducedGraph with the reduced InducedGraph
	newPGraph := PermanentInducedGraph{
		IGraph:            *rGraph,
		PermanentVertices: pGraph.PermanentVertices.Copy(),
	}

	if v1.Id >= 0 {
		_ = newPGraph.PermanentVertices.AddVertex(v1)
		// this is the only change
		// if a new vertex is added, we check, going from it, which can be marked as bad
		// we only need to do this from the currently added vertex, because the search from the
		// others has already been done and the results can not change.
		err := pGraph.MarkBadWithBfs(vToRemove, 2)
		if err != nil {
			return nil, err
		}
	}
	if v2.Id >= 0 {
		_ = newPGraph.PermanentVertices.AddVertex(v2)
		// this is the only change to the other function CreateReduced
		err := pGraph.MarkBadWithBfs(vToRemove, 2)
		if err != nil {
			return nil, err
		}
	}

	return &newPGraph, nil
}
