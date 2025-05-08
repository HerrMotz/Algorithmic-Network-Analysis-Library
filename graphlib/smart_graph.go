package graphlib

// This library makes a set of graph data structures available through
// a common interface. The goal is to provide a simple way to
// create and manipulate graphs while keeping the implementation
// details hidden.

type Graph interface {
	// AddEdge adds an edge between two vertices.
	AddEdge(u, v int) error
	// RemoveEdge removes an edge between two vertices.
	RemoveEdge(u, v int) error
	// Neighbours returns the neighbours of a vertex.
	Neighbours(u int) ([]int, error)
	// NeighboursCount returns the number of neighbours to a vertex.
	NeighboursCount(u int) (int, error)
	// Size returns the number of vertices in the graph.
	Size() int
	// HasEdge returns true if there is an edge between two vertices.
	HasEdge(u, v int) (bool, error)
	// AddVertex adds a vertex to the graph.
	AddVertex() (int, error)
	// RemoveVertex removes a vertex from the graph.
	RemoveVertex(u int) error
	// HasVertex returns true if a vertex exists in the graph.
	HasVertex(u int) (bool, error)
	// Clear removes all edges and vertices from the graph.
	Clear() error
	// String returns a string representation of the graph.
	String() string
	// Clone creates a deep copy of the graph.
	Clone() (Graph, error)
	// MarshalJSON serializes the graph to JSON.
	MarshalJSON() ([]byte, error)
	// UnmarshalJSON deserializes the graph from JSON.
	UnmarshalJSON(data []byte) error
}
