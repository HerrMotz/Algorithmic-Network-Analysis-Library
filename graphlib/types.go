package graphlib

import (
	"errors"
	"math"
)

type Arc struct {
	From Vertex `json:"from,omitempty"`
	To   Vertex `json:"to,omitempty"`
}
type Edge struct {
	U Vertex `json:"u,omitempty"`
	V Vertex `json:"v,omitempty"`
}

type Vertex struct {
	Id int
}

const INF = math.MaxUint32 // or MaxInt64, depending on your needs

var EdgeOrArcExistsError = errors.New("edge or arc already exists")

type Graph interface {
	Size() int
	HasEdge(edge Edge) bool
	HasArc(arc Arc) bool
	Neighbours(u Vertex) ([]int, error)
	Degree(u Vertex) (int, error)
	AddEdge(edge Edge) error
	AddArc(arc Arc) error
	RemoveEdge(edge Edge) error
	RemoveArc(arc Arc) error
}
