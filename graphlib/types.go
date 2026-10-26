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

// EdgeOrArcNotFoundError is returned when an edge or arc that should be removed
// does not exist.
var EdgeOrArcNotFoundError = errors.New("edge or arc does not exist")
