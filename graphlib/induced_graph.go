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

func (igraph *InducedGraph) Size() int {
	return igraph.GoodVertices.Size()
}

func (igraph *InducedGraph) RemoveVertex(v Vertex) error {
	return igraph.MarkVertexAsBad(v)
}

func (igraph *InducedGraph) MarkVertexAsBad(v Vertex) error {
	// marks vertex v as bad in the induced graph
	if igraph.AdjList == nil || igraph.GoodVertices == nil {
		return errors.New("induced graph is not initialized")
	}
	if !igraph.GoodVertices.HasVertex(v) {
		return errors.New(fmt.Sprintf("Vertex %v does not exist", v))
	}

	// Remove vertex from GoodVertices
	igraph.GoodVertices.RemoveVertex(v)

	// Remove vertex from AdjList
	igraph.AdjList = igraph.AdjList.RemoveVertexCopy(v)
	return nil
}

func (igraph *InducedGraph) Neighbours(v Vertex) (VertexSet, error) {
	// returns the neighbours of vertex v in the induced graph
	if igraph.AdjList == nil || igraph.GoodVertices == nil {
		return nil, errors.New("induced graph is not initialized")
	}
	if !igraph.GoodVertices.HasVertex(v) {
		return nil, errors.New(fmt.Sprintf("Vertex %v does not exist", v))
	}

	neighbours, err := igraph.AdjList.Neighbours(v)
	if err != nil {
		return nil, err
	}
	// copy the neighbours to avoid modifying the original set
	neighbours = neighbours.Copy()

	// Filter neighbours to only include those that are in GoodVertices
	if igraph.GoodVertices.Size() < igraph.AdjList.Size { // only check if node set is smaller than the original graph
		for k := range neighbours {
			if !igraph.GoodVertices.HasVertex(Vertex{k}) {
				neighbours.RemoveVertex(Vertex{k})
			}
		}
	}
	return neighbours, nil
}

func (igraph *InducedGraph) HasVertex(v Vertex) bool {
	if igraph.GoodVertices == nil {
		return false
	}
	return igraph.GoodVertices.HasVertex(v)
}

func (igraph *InducedGraph) Degree(v Vertex) (int, error) {
	// returns the degree of vertex v in the induced graph
	if igraph.AdjList == nil || igraph.GoodVertices == nil {
		return 0, errors.New("induced graph is not initialized")
	}
	if !igraph.GoodVertices.HasVertex(v) {
		return 0, errors.New(fmt.Sprintf("Vertex %v does not exist", v))
	}

	neighbours, err := igraph.Neighbours(v)
	if err != nil {
		return 0, err
	}
	return len(neighbours), nil
}

// CreateReduced returns a new InducedGraph that shares the same
// AdjList as igraph but whose GoodVertices set has been copied, filtered
// and with vToRemove removed. The original igraph is not modified.
func (igraph *InducedGraph) CreateReduced(
	vToRemove Vertex,
) (*InducedGraph, error) {
	if igraph.AdjList == nil || igraph.GoodVertices == nil {
		return nil, errors.New("induced graph is not initialized")
	}
	if !igraph.GoodVertices.HasVertex(vToRemove) {
		return nil, fmt.Errorf("vertex %v does not exist in good vertices", vToRemove)
	}

	// Copy the GoodVertices set and remove the specified vertex
	newGood := igraph.GoodVertices.CopyAndRemoveVertex(vToRemove)

	// Return a brand-new InducedGraph sharing the same AdjList
	return &InducedGraph{
		AdjList:      igraph.AdjList,
		GoodVertices: newGood,
	}, nil
}

// CreateReduced combines the necessary steps of PGraph and IGraph to create a new
// PermanentInducedGraph that shares the same AdjList as pgraph.IGraph but
// where vToRemove is removed from the GoodVertices set.
func (pgraph *PermanentInducedGraph) CreateReduced(vToRemove, v1, v2 Vertex) (*PermanentInducedGraph, error) {
	if pgraph.IGraph.AdjList == nil || pgraph.IGraph.GoodVertices == nil {
		return nil, errors.New("permanent induced graph is not initialized")
	}

	rGraph, err := pgraph.IGraph.CreateReduced(vToRemove)
	if err != nil {
		return nil, fmt.Errorf("failed to create reduced graph: %w", err)
	}

	// Create a new PermanentInducedGraph with the reduced InducedGraph
	newPGraph := PermanentInducedGraph{
		IGraph:            *rGraph,
		PermanentVertices: pgraph.PermanentVertices.Copy(),
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
